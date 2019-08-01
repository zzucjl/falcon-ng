package model

import (
	"fmt"
)

type NodeEndpoint struct {
	NodeId     int64 `xorm:"'node_id'"`
	EndpointId int64 `xorm:"'endpoint_id'"`
}

func (NodeEndpoint) TableName() string {
	return "node_endpoint"
}

func NodeIdsGetByEndpointId(endpointId int64) ([]int64, error) {
	if endpointId == 0 {
		return []int64{}, nil
	}

	var ids []int64
	err := DB["portal"].Table("node_endpoint").Where("endpoint_id = ?", endpointId).Select("node_id").Find(&ids)
	return ids, err
}

func EndpointIdsByNodeIds(nodeIds []int64) ([]int64, error) {
	if nodeIds == nil || len(nodeIds) == 0 {
		return []int64{}, nil
	}

	var ids []int64
	err := DB["portal"].Table("node_endpoint").In("node_id", nodeIds).Select("endpoint_id").Find(&ids)
	return ids, err
}

func NodeEndpointGetByEndpointIds(endpointsIds []int64) ([]NodeEndpoint, error) {
	if endpointsIds == nil || len(endpointsIds) == 0 {
		return []NodeEndpoint{}, nil
	}

	var objs []NodeEndpoint
	err := DB["portal"].In("endpoint_id", endpointsIds).Find(&objs)
	return objs, err
}

func NodeEndpointGetByNodeIds(nodeIds []int64) ([]NodeEndpoint, error) {
	if nodeIds == nil || len(nodeIds) == 0 {
		return []NodeEndpoint{}, nil
	}

	var objs []NodeEndpoint
	err := DB["portal"].In("node_id", nodeIds).Find(&objs)
	return objs, err
}

func NodeEndpointUnbind(nid, eid int64) error {
	_, err := DB["portal"].Where("node_id=? and endpoint_id=?", nid, eid).Delete(new(NodeEndpoint))
	return err
}

func NodeEndpointBind(nid, eid int64) error {
	total, err := DB["portal"].Where("node_id=? and endpoint_id=?", nid, eid).Count(new(NodeEndpoint))
	if err != nil {
		return err
	}

	if total > 0 {
		return nil
	}

	endpoint, err := EndpointGet("id", eid)
	if err != nil {
		return err
	}

	if endpoint == nil {
		return fmt.Errorf("endpoint[id:%d] not found", eid)
	}

	_, err = DB["portal"].Insert(&NodeEndpoint{
		NodeId:     nid,
		EndpointId: eid,
	})

	return err
}
