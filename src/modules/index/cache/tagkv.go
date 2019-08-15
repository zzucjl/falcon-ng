package cache

import (
	"errors"
	"fmt"
	"sort"

	"github.com/toolkits/pkg/logger"
)

type TagkvStruct struct {
	TagK string   `json:"tagk"`
	TagV []string `json:"tagv"`
}

type XCludeList []*TagkvStruct

// <-- 100/tag1=v1,tag2=v2
func (x *XCludeList) Include(tagMap map[string]string) bool {
	if len(tagMap) == 0 {
		return false
	}

	for _, cludeStruct := range *x {
		ctagV, ok := tagMap[cludeStruct.TagK]
		if !ok {
			return false
		}

		find := false
		for _, tagV := range cludeStruct.TagV {
			if ctagV == tagV {
				find = true
				break
			}
		}

		if !find {
			return false
		}
	}

	return true
}

func (x *XCludeList) Exclude(tagMap map[string]string) bool {
	if len(tagMap) == 0 {
		return true
	}

	for _, cludeStruct := range *x {
		ctagV, ok := tagMap[cludeStruct.TagK]
		if ok {
			for _, tagV := range cludeStruct.TagV {
				if tagV == ctagV {
					return false
				}
			}
		}
	}

	return true
}

func (x *XCludeList) GetAllCombinationString() ([]string, error) {
	listLen := len(*x)
	newTags := make(XCludeList, listLen)
	tagsMap := make(map[string][]string)
	keys := make([]string, listLen)
	i := 0
	for _, xcludeStruct := range *x {
		keys[i] = xcludeStruct.TagK
		tagsMap[xcludeStruct.TagK] = xcludeStruct.TagV
		i++
	}

	// check是否有相同的TagK
	if len(keys) != len(tagsMap) {
		return []string{}, errors.New("the tagName must be unique")
	}

	sort.Strings(keys)

	for j, key := range keys {
		newTags[j] = &TagkvStruct{TagK: key, TagV: tagsMap[key]}
	}
	return x.getAllCombinationComplex(newTags), nil
}

func (x *XCludeList) getAllCombinationComplex(tags XCludeList) []string {
	if len(tags) == 0 {
		return []string{}
	}
	firstStruct := tags[0]
	firstList := make([]string, len(firstStruct.TagV))

	for i, v := range firstStruct.TagV {
		// firstList[i] = fmt.Sprintf("%s=%s", firstStruct.TagK, v)
		firstList[i] = firstStruct.TagK + "=" + v
	}

	otherList := x.getAllCombinationComplex(tags[1:])
	if len(otherList) == 0 {
		return firstList
	} else {
		toAlloc := len(otherList) * len(firstList)
		// hard code, 100W限制
		if toAlloc >= 1000000 {
			logger.Warningf("getAllCombinationComplex try to makearray with size:%d", toAlloc)
			logger.Warningf("getAllCombinationComplex input: %v", x)
		}
		retList := make([]string, len(otherList)*len(firstList))
		i := 0
		for _, firstV := range firstList {
			for _, otherV := range otherList {
				// retList[i] = fmt.Sprintf("%s,%s", firstV, otherV)
				retList[i] = firstV + "," + otherV
				i++
			}
		}

		return retList
	}
}

//Check if can over limit
func (x *XCludeList) CheckFullMatch(limit int64) error {
	multiRes := int64(1)

	for _, xStruct := range *x {
		multiRes = multiRes * int64(len(xStruct.TagV))
		if multiRes > limit {
			return fmt.Errorf("err too many tags")
		}
	}

	if multiRes == 0 {
		return fmt.Errorf("err empty tagk")
	}

	return nil
}
