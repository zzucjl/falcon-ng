package cron

import (
	"fmt"
	"strings"
	"time"

	"github.com/toolkits/pkg/logger"

	"github.com/open-falcon/falcon-ng/src/model"
	"github.com/open-falcon/falcon-ng/src/modules/alarm/cache"
	"github.com/open-falcon/falcon-ng/src/modules/alarm/config"
	"github.com/open-falcon/falcon-ng/src/modules/alarm/notify"
	"github.com/open-falcon/falcon-ng/src/modules/alarm/redi"
)

const RECOVERY_TIME_PREFIX = "/falcon-ng/recovery/time/"
const ALERT_TIME_PREFIX = "/falcon-ng/alert/time/"
const ALERT_UPGRADE_PREFIX = "/falcon-ng/alert/upgrade/"

func consume(event *model.Event, isHigh bool) {
	if event == nil {
		return
	}

	if IsMaskEvent(event) {
		SetEventStatus(event, model.STATUS_MASK)
		return
	}

	if event.NeedUpgrade == 1 {
		needUpgrade, needNotify := isAlertUpgrade(event)
		if needUpgrade {
			var alertUpgrade model.EventAlertUpgrade
			if err := json.Unmarshal([]byte(event.AlertUpgrade), &alertUpgrade); err != nil {
				logger.Errorf("AlertUpgrade unmarshal failed, event: %+v, err: %v", event, err)
				return
			}

			if event.EventType == config.ALERT {
				err := model.UpdateEventCurPriority(event.HashId, alertUpgrade.Level)
				if err != nil {
					logger.Errorf("UpdateEventCurPriority failed, err: %v, event: %+v", err, event)
					return
				}
			}
			err := model.UpdateEventPriority(event.Id, alertUpgrade.Level)
			if err != nil {
				logger.Errorf("UpdateEventPriority failed, err: %v, event: %+v", err, event)
				return
			}
			event.Priority = alertUpgrade.Level

			SetEventStatus(event, model.STATUS_UPGRADE)

			if needNotify {
				if event.EventType == config.ALERT && NeedCallback(event.Sid) {
					if err := PushCallbackEvent(event); err != nil {
						logger.Errorf("push event to callback queue failed, callbackEvent: %+v", event)
					}
					logger.Infof("push event to callback queue succ, event: %+v", event)

					SetEventStatus(event, model.STATUS_CALLBACK)
				}

				go notify.DoNotify(true, event)
				SetEventStatus(event, model.STATUS_SEND)
				return
			}

			SetEventStatus(event, model.STATUS_CONVERGE)
			return
		}
	}

	if isInConverge(event, false) {
		SetEventStatus(event, model.STATUS_CONVERGE)
		return
	}

	if event.EventType == config.ALERT && NeedCallback(event.Sid) {
		if err := PushCallbackEvent(event); err != nil {
			logger.Errorf("push event to callback queue failed, callbackEvent: %+v", event)
		}
		logger.Infof("push event to callback queue succ, event: %+v", event)

		SetEventStatus(event, model.STATUS_CALLBACK)
	}

	// 没有配置报警接收人，修改event状态为无接收人
	if strings.TrimSpace(event.Users) == "[]" && strings.TrimSpace(event.Groups) == "[]" {
		SetEventStatus(event, model.STATUS_NONEUSER)
		return
	}

	if !isHigh {
		storeLowEvent(event)
		return
	}

	go notify.DoNotify(false, event)
	SetEventStatus(event, model.STATUS_SEND)
}

// isInConverge 包含2种情况
// 1. 用户配置了N秒之内只报警M次
// 2. 用户配置了不发送recovery报警
func isInConverge(event *model.Event, isUpgrade bool) bool {
	stra, exists := cache.StraCache.GetById(event.Sid)
	if !exists {
		logger.Errorf("sid not found, event: %+v", event)
		return false
	}

	eventString := RECOVERY_TIME_PREFIX + fmt.Sprint(event.HashId)

	now := time.Now().Unix()

	if event.EventType == config.RECOVERY {
		redi.SetWithTTL(eventString, now, 30*24*3600)
		if stra.RecoveryNotify == 0 {
			return true
		}

		return false
	}

	convergeInSeconds := int64(stra.Converge[0])
	convergeMaxCounts := int64(stra.Converge[1])

	// 最多报0次，收敛该报警
	if convergeMaxCounts == 0 {
		return true
	}

	// 无收敛
	if convergeInSeconds == 0 {
		return false
	}

	var recoveryTs int64
	if redi.HasKey(eventString) {
		recoveryTs = redi.GET(eventString)
	}

	// 计算起始时间戳, 不能用event time来计算，因为从redis消费的时候是从队尾拿的数据
	startTs := now - convergeInSeconds
	if startTs < recoveryTs {
		startTs = recoveryTs
	}

	cnt, err := model.EventCnt(event.HashId, model.ParseEtime(startTs), model.ParseEtime(now), isUpgrade)
	if err != nil {
		logger.Errorf("get event count failed, err: %v", err)
		return false
	}

	if cnt >= convergeMaxCounts {
		logger.Infof("converge max counts: %c reached, currend: %v, event: %+v", convergeMaxCounts, cnt, event)
		return true
	}

	return false
}

// 三种情况，不需要升级报警
// 1，被认领的报警不需要升级
// 2，忽略的报警不需要升级
// 3，屏蔽的报警，不需要升级
func isAlertUpgrade(event *model.Event) (needUpgrade, needNotify bool) {
	alertUpgradeKey := ALERT_UPGRADE_PREFIX + fmt.Sprint(event.HashId)
	eventAlertKey := ALERT_TIME_PREFIX + fmt.Sprint(event.HashId)

	if event.EventType == config.RECOVERY {
		err := model.EventCurDel(event.HashId)
		if err != nil {
			logger.Errorf("del event cur failed, event: %+v, err: %v", event, err)
		}

		if redi.HasKey(alertUpgradeKey) {
			err := redi.DelKey(eventAlertKey)
			if err != nil {
				logger.Errorf("redis del eventAlertkey failed, key: %v, err: %v", eventAlertKey, err)
			}

			err = redi.DelKey(alertUpgradeKey)
			if err != nil {
				logger.Errorf("redis del alertUpgradeKey failed, key: %v, err: %v", alertUpgradeKey, err)
			}

			return true, true
		}

		return false, false
	}

	eventCur, err := model.EventCurGet("hashid", event.HashId)
	if err != nil {
		logger.Errorf("AlertUpgrade failed:get event_cur failed, event: %+v, err: %v", event, err)
		return false, false
	}

	if eventCur == nil {
		logger.Infof("AlertUpgrade failed:get event_cur is nil, event: %+v", event)
		return false, false
	}

	now := time.Now().Unix()

	var alertUpgrade model.EventAlertUpgrade
	if err = json.Unmarshal([]byte(event.AlertUpgrade), &alertUpgrade); err != nil {
		logger.Errorf("AlertUpgrade unmarshal failed, event: %+v, err: %v", event, err)
		return false, false
	}

	upgradeDuration := int64(alertUpgrade.Duration)

	claimants := strings.TrimSpace(eventCur.Claimants)
	if claimants != "[]" && claimants != "" {
		return false, false
	}

	if eventCur.IgnoreAlert == 1 {
		return false, false
	}

	if !redi.HasKey(eventAlertKey) {
		err := redi.SetWithTTL(eventAlertKey, now, 30*24*3600)
		if err != nil {
			logger.Errorf("set eventAlertKey failed, eventAlertKey: %v, err: %v", eventAlertKey, err)
			return false, false
		}
	}

	firstAlertTime := redi.GET(eventAlertKey)
	if now-firstAlertTime < upgradeDuration {
		return false, false
	}

	err = redi.SetWithTTL(alertUpgradeKey, 1, 30*24*3600)
	if err != nil {
		logger.Errorf("set alertUpgradeKey failed, alertUpgradeKey: %v, err: %v", alertUpgradeKey, err)
		return false, false
	}

	if isInConverge(event, true) {
		return true, false
	}

	return true, true
}

func SetEventStatus(event *model.Event, status string) {
	if err := model.SaveEventStatus(event.Id, status); err != nil {
		logger.Errorf("set event status failed, event: %+v, status: %v, err:%v", event, status, err)
	} else {
		logger.Infof("set event status succ, event: %+v, status: %v", event, status)
	}

	if event.EventType == config.ALERT {
		if err := model.SaveEventCurStatus(event.HashId, status); err != nil {
			logger.Errorf("set event_cur status failed, event: %+v, status: %v, err:%v", event, status, err)
		} else {
			logger.Infof("set event_cur status succ, event: %+v, status: %v", event, status)
		}
	}
}
