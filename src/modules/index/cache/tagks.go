package cache

import (
	"sync"
)

//TagKeys
type TagksStruct struct { // ns/metric -> tagk
	sync.RWMutex
	Tagks map[string]*TagkStruct `json:"tagks"`
}

func (t TagksStruct) New() *TagksStruct {
	return &TagksStruct{
		Tagks: make(map[string]*TagkStruct, 0),
	}
}

func (t *TagksStruct) Clean(now, timeDuration int64) {
	t.Lock()
	defer t.Unlock()

	for tagk, tagkStruct := range t.Tagks {
		if now-tagkStruct.Updated > timeDuration {
			delete(t.Tagks, tagk)
		} else {
			//清理tagvs
			tagkStruct.Tagvs.Clean(now, timeDuration)
		}
	}
}

func (t *TagksStruct) CleanEndpoint(endpoint string) {
	t.Lock()
	defer t.Unlock()
	for tagk, tagkStruct := range t.Tagks {
		if tagk == "endpoint" {
			tagkStruct.Tagvs.CleanEndpoint(endpoint)

			return
		}
	}
}

func (t *TagksStruct) GetTagkv() []*TagkvStruct {
	t.RLock()
	defer t.RUnlock()
	tagkvs := []*TagkvStruct{}
	for k, tagvs := range t.Tagks {
		vs := tagvs.Tagvs.GetTagvs()
		tagkv := TagkvStruct{
			TagK: k,
			TagV: vs,
		}
		tagkvs = append(tagkvs, &tagkv)
	}

	return tagkvs
}

func (t *TagksStruct) GetTagkvMap() map[string][]string {
	t.RLock()
	defer t.RUnlock()
	tagkvs := make(map[string][]string)

	for k, tagvs := range t.Tagks {
		vs := tagvs.Tagvs.GetTagvs()
		tagkvs[k] = vs
	}

	return tagkvs
}

func (t *TagksStruct) MustGetTagkStruct(k string, now int64) *TagkStruct {
	t.Lock()
	defer t.Unlock()
	if _, exists := t.Tagks[k]; !exists {
		t.Tagks[k] = NewTagkStruct(now)
	}
	return t.Tagks[k]
}

//TagKey
type TagkStruct struct {
	Updated int64

	Tagvs *TagkTagvsStruct
}

func NewTagkStruct(ts int64) *TagkStruct {
	return &TagkStruct{Updated: ts, Tagvs: NewTagkTagvsStruct()}
}
