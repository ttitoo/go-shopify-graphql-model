package model

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/mitchellh/mapstructure"
)

// UnmarshalJSON implements the json.Unmarshaler interface.
func (c *WebhookSubscriptionConnection) UnmarshalJSON(b []byte) error {
	var m map[string]interface{}
	err := json.Unmarshal(b, &m)
	if err != nil {
		return err
	}

	data := m["webhookSubscriptions"].(map[string]interface{})
	if edges, ok := data["edges"].([]interface{}); ok {
		if edges != nil && len(edges) > 0 {
			c.Edges = make([]WebhookSubscriptionEdge, len(edges))
			for i, e := range edges {
				edge, err := decodeWebhookSubscriptionEdge(e.(map[string]interface{}))
				if err != nil {
					return err
				}

				c.Edges[i] = *edge
			}
		}
	}

	if nodes, ok := data["nodes"].([]interface{}); ok {
		if len(nodes) > 0 {
			c.Nodes = make([]WebhookSubscription, len(nodes))
			for j, e := range nodes {
				node, err := decodeWebhookSubscription(e.(map[string]interface{}))
				if err != nil {
					return err
				}

				c.Nodes[j] = *node
			}
		}
	}
	return nil
}

func decodeWebhookSubscription(m map[string]interface{}) (*WebhookSubscription, error) {
	node := WebhookSubscription{}
	endpointData := m["endpoint"].(map[string]interface{})
	delete(m, "endpoint")

	err := mapstructure.Decode(m, &node)
	if err != nil {
		return nil, fmt.Errorf("decode webhook subscription node: %w", err)
	}

	var typename string
	if val, ok := endpointData["__typename"].(string); ok {
		typename = val
	} else if val, ok := m["id"].(string); ok {
		submatches := gidRegex.FindStringSubmatch(val)
		if len(submatches) != 2 {
			return nil, fmt.Errorf("malformed gid=`%s`", val)
		}
		typename = submatches[1]
	} else {
		return nil, fmt.Errorf("can not detect WebhookSubscriptionEndpoint")
	}

	t, err := detectEndpointType(typename)
	if err != nil {
		return nil, err
	}

	endpoint := reflect.New(t).Interface()
	err = mapstructure.Decode(endpointData, &endpoint)
	if err != nil {
		return nil, fmt.Errorf("decode webhook subscription endpoint node: %w", err)
	}

	node.Endpoint = endpoint.(WebhookSubscriptionEndpoint)
	return &node, nil
}

func decodeWebhookSubscriptionEdge(m map[string]interface{}) (*WebhookSubscriptionEdge, error) {
	s, err := decodeWebhookSubscription(m["node"].(map[string]interface{}))
	if err != nil {
		return nil, err
	}

	var cursor string
	if val, ok := m["cursor"]; ok {
		cursor = val.(string)
	}

	return &WebhookSubscriptionEdge{
		Cursor: cursor,
		Node:   s,
	}, nil
}

func (s *MediaEdge) UnmarshalJSON(b []byte) error {
	var m map[string]interface{}
	err := json.Unmarshal(b, &m)
	if err != nil {
		return err
	}
	if cursor, ok := m["cursor"].(string); ok {
		s.Cursor = cursor
	}
	if node, ok := m["node"].(map[string]interface{}); ok {
		s.Node, err = decodeMedia(node)
		if err != nil {
			return fmt.Errorf("decode media node: %w", err)
		}
	}
	return nil
}

func (s *MediaConnection) UnmarshalJSON(b []byte) error {
	var (
		m     map[string]interface{}
		mConn struct {
			Edges    []MediaEdge `json:"edges,omitempty"`
			PageInfo *PageInfo   `json:"pageInfo,omitempty"`
		}
	)
	err := json.Unmarshal(b, &mConn)
	if err != nil {
		return err
	}
	s.Edges = mConn.Edges
	s.PageInfo = mConn.PageInfo

	err = json.Unmarshal(b, &m)
	if err != nil {
		return err
	}
	if nodes, ok := m["nodes"].([]interface{}); ok {
		s.Nodes = make([]Media, len(nodes))
		for i, n := range nodes {
			if node, ok := n.(map[string]interface{}); ok {
				s.Nodes[i], err = decodeMedia(node)
				if err != nil {
					return fmt.Errorf("decode media node: %w", err)
				}
			} else {
				return fmt.Errorf("expected type map[string]interface{} for Media node, got %T", n)
			}
		}
	}
	return nil
}

func decodeMedia(node map[string]interface{}) (Media, error) {
	if id, ok := node["id"].(string); ok {
		mediaType, err := concludeObjectType(id)
		if err != nil {
			return nil, fmt.Errorf("conclude object type: %w", err)
		}
		media := reflect.New(mediaType).Interface()
		err = mapstructure.Decode(node, media)
		if err != nil {
			return nil, fmt.Errorf("decode media node: %w", err)
		}
		return media.(Media), nil
	}
	return nil, fmt.Errorf("must query id to decode Media")
}
