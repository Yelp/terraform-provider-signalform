package signalform

import (
	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetNotifications(t *testing.T) {
	values := []interface{}{
		"Email,test@yelp.com",
		"PagerDuty,credId",
		"Webhook,test,https://foo.bar.com?user=test&action=alert",
	}

	expected := []map[string]interface{}{
		map[string]interface{}{
			"type":  "Email",
			"email": "test@yelp.com",
		},
		map[string]interface{}{
			"type":         "PagerDuty",
			"credentialId": "credId",
		},
		map[string]interface{}{
			"type":   "Webhook",
			"secret": "test",
			"url":    "https://foo.bar.com?user=test&action=alert",
		},
	}
	assert.Equal(t, expected, getNotifications(values))
}

func TestResourceRuleHash(t *testing.T) {
	// Tests basic and consistent hashing, keys in the maps are sorted
	values := map[string]interface{}{
		"description":  "Test Rule Name",
		"detect_label": "Test Detect Label",
		"severity":     "Critical",
		"disabled":     "true",
	}

	expected := hashcode.String("Test Rule Name-Critical-Test Detect Label-true-")
	assert.Equal(t, expected, resourceRuleHash(values))

	// Test new params in rules
	values = map[string]interface{}{
		"description":  "Test Rule Name",
		"detect_label": "Test Detect Label",
		"severity":     "Critical",
		"disabled":     "true",
		"parameterized_subject": "Test subject",
		"parameterized_body": "Test body",
	}

	expected = hashcode.String("Test Rule Name-Critical-Test Detect Label-true-Test body-Test subject-")
	assert.Equal(t, expected, resourceRuleHash(values))

	values = map[string]interface{}{
		"description":  "Test Rule Name",
		"detect_label": "Test Detect Label",
		"severity":     "Critical",
		"disabled":     "true",
		"parameterized_subject": "Test subject",
		"parameterized_body": "Test body",
		"runbook_url": "https://example.com",
		"tip": "test tip",
	}

	expected = hashcode.String("Test Rule Name-Critical-Test Detect Label-true-Test body-Test subject-https://example.com-test tip-")
	assert.Equal(t, expected, resourceRuleHash(values))
}

func TestValidateSeverityAllowed(t *testing.T) {
	_, errors := validateSeverity("Critical", "severity")
	assert.Equal(t, len(errors), 0)
}

func TestValidateSeverityNotAllowed(t *testing.T) {
	_, errors := validateSeverity("foo", "severity")
	assert.Equal(t, len(errors), 1)
}
