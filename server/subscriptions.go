package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
)

// Subscription Object
type Subscription struct {
	ChannelID string
	URL       string
	XMLInfo   *XMLFeedMeta
}

// XMLFeedMeta stores part of the feed xml data
type XMLFeedMeta struct {
	ID, Title, GUID, PubDate string `json:",omitempty"`
}

const SUBSCRIPTIONS_KEY = "subscriptions"

// Subscriptions map to key value pairs
type Subscriptions struct {
	Subscriptions map[string]*Subscription
}

// Subscribe prosses the /feed subscribe <channel> <url>
func (p *RSSFeedPlugin) subscribe(ctx context.Context, channelID string, url string) error {
	sub := &Subscription{
		ChannelID: channelID,
		URL:       url,
		XMLInfo:   &XMLFeedMeta{},
	}

	key := getKey(channelID, url)
	if err := p.addSubscription(key, sub); err != nil {
		p.API.LogError(err.Error())
		return err
	}

	return nil
}

func (p *RSSFeedPlugin) addSubscription(key string, sub *Subscription) error {
	currentSubscriptions, err := p.getSubscriptions()
	if err != nil {
		p.API.LogError(err.Error())
		return err
	}

	// check if url already exists
	_, ok := currentSubscriptions.Subscriptions[key]
	if !ok {
		currentSubscriptions.Subscriptions[key] = sub
		err = p.storeSubscriptions(currentSubscriptions)
		if err != nil {
			p.API.LogError(err.Error())
			return err
		}

	}

	return nil
}

func (p *RSSFeedPlugin) getSubscriptions() (*Subscriptions, error) {
	var subscriptions *Subscriptions

	value, err := p.API.KVGet(SUBSCRIPTIONS_KEY)
	if err != nil {
		p.API.LogError(err.Error())
		return nil, err
	}

	if value == nil {
		subscriptions = &Subscriptions{Subscriptions: map[string]*Subscription{}}
	} else {
		json.NewDecoder(bytes.NewReader(value)).Decode(&subscriptions)
	}

	return subscriptions, nil
}

func (p *RSSFeedPlugin) storeSubscriptions(s *Subscriptions) error {
	b, err := json.Marshal(s)
	if err != nil {
		p.API.LogError(err.Error())
		return err
	}

	p.API.KVSet(SUBSCRIPTIONS_KEY, b)
	return nil
}

func (p *RSSFeedPlugin) unsubscribe(channelID string, url string) error {

	currentSubscriptions, err := p.getSubscriptions()
	if err != nil {
		p.API.LogError(err.Error())
		return err
	}

	key := getKey(channelID, url)
	_, ok := currentSubscriptions.Subscriptions[key]
	if ok {
		delete(currentSubscriptions.Subscriptions, key)
		if err := p.storeSubscriptions(currentSubscriptions); err != nil {
			p.API.LogError(err.Error())
			return err
		}
	}

	return nil
}

func (p *RSSFeedPlugin) updateSubscription(subscription *Subscription) error {
	currentSubscriptions, err := p.getSubscriptions()
	if err != nil {
		p.API.LogError(err.Error())
		return err
	}

	key := getKey(subscription.ChannelID, subscription.URL)
	_, ok := currentSubscriptions.Subscriptions[key]
	if ok {
		currentSubscriptions.Subscriptions[key] = subscription
		if err := p.storeSubscriptions(currentSubscriptions); err != nil {
			p.API.LogError(err.Error())
			return err
		}
	}
	return nil
}

func getKey(channelID string, url string) string {
	return fmt.Sprintf("%s/%s", channelID, url)
}
