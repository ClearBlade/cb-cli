package mqttclient

import (
	//	"fmt"
	"strings"
)

func (ot *outgoing_topics) bestTopicMatch(topic string) string {

	actual := strings.Split(topic, "/")
	bestMatch := ""
	bestScore := -99999

	for subscription := range ot.store {
		sub := strings.Split(subscription, "/")
		if !isValidSubscriptionTopic(sub) {
			//fmt.Printf("Bad subscription topic: %s\n", subscription)
			continue
		}
		matched := matchActualToSubscription(actual, sub)
		if matched {
			if curScore := scoreSubscription(sub, 0); curScore > bestScore {
				bestMatch = subscription
				bestScore = curScore
			}
			//fmt.Printf("Matched: %s %d\n", subscription, scoreSubscription(sub, 0))
		} else {
			//fmt.Printf("Did not match: %s\n", subscription)
		}
	}
	if bestMatch == "" {
		//fmt.Printf("No match for %s\n", topic)
		bestMatch = topic
	}
	return bestMatch
}

func matchActualToSubscription(actual, subscription []string) bool {
	if !(len(actual) == len(subscription) || (len(actual) > len(subscription) &&
		lastElementIsPound(subscription))) {
		return false
	}
	return matchTopic(actual, subscription)
}

func matchTopic(actual, subscription []string) bool {
	if len(actual) == 0 || (len(subscription) == 1 && subscription[0] == "#") {
		return true
	}
	if subscription[0] == actual[0] || subscription[0] == "+" {
		return matchTopic(actual[1:], subscription[1:])
	}
	return false
}

func isValidSubscriptionTopic(topic []string) bool {
	if len(topic) == 0 {
		return true
	}
	if len(topic) > 1 && topic[0] == "#" {
		return false
	}
	return isValidSubscriptionTopic(topic[1:])
}

func lastElementIsPound(topic []string) bool {
	return len(topic) > 0 && topic[len(topic)-1] == "#"
}

func scoreSubscription(subscription []string, currentScore int) int {
	if len(subscription) == 0 {
		return currentScore
	}
	if subscription[0] == "#" {
		currentScore = currentScore - 2
	} else if subscription[0] == "+" {
		currentScore = currentScore - 1
	}
	return scoreSubscription(subscription[1:], currentScore+1)
}
