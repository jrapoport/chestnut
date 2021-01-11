package tags

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseJSONTag(t *testing.T) {
	tests := []struct {
		tag  string
		name string
		opts []string
	}{
		{"", "", []string{}},
		{"-", "-", []string{}},
		{"test", "test", []string{}},
		{",secure", "", []string{SecureOption}},
		{"-,secure", "-", []string{}},
		{"test,secure", "test", []string{SecureOption}},
		{",secure,hash", "", []string{SecureOption, HashOption}},
		{"-,secure,hash", "-", []string{}},
		{"test,secure,hash", "test", []string{SecureOption, HashOption}},
		{",secure,hash,omitempty", "", []string{SecureOption, HashOption, "omitempty"}},
		{"-,secure,hash,omitempty", "-", []string{}},
		{"test,secure,hash,omitempty", "test", []string{SecureOption, HashOption, "omitempty"}},
	}
	for _, test := range tests {
		name, opts := ParseJSONTag(test.tag)
		assert.Equal(t, test.name, name)
		assert.ElementsMatch(t, test.opts, opts)
	}
}

func TestIgnoreField(t *testing.T) {
	tests := []struct {
		tag        string
		assertBool assert.BoolAssertionFunc
	}{
		{"", assert.False},
		{"-", assert.True},
		{"test", assert.False},
		{",secure", assert.False},
		{"-,secure", assert.True},
		{"test,secure", assert.False},
		{",secure,hash", assert.False},
		{"-,secure,hash", assert.True},
		{"test,secure,hash", assert.False},
		{",secure,hash,omitempty", assert.False},
		{"-,secure,hash,omitempty", assert.True},
		{"test,secure,hash,omitempty", assert.False},
	}
	for _, test := range tests {
		name, _ := ParseJSONTag(test.tag)
		test.assertBool(t, IgnoreField(name), "unexpected")
	}
}

func TestHasOption(t *testing.T) {
	tests := []struct {
		opts []string
		opt  string
		has  bool
	}{
		{[]string{}, "", false},
		{[]string{}, SecureOption, false},
		{[]string{HashOption}, SecureOption, false},
		{[]string{SecureOption}, SecureOption, true},
		{nil, SecureOption, false},
	}
	for _, test := range tests {
		has := HasOption(test.opts, test.opt)
		assert.Equal(t, test.has, has)
	}
}

func TestHashFunction(t *testing.T) {
	tests := []struct {
		opts []string
		name Hash
	}{
		{nil, HashNone},
		{[]string{}, HashNone},
		{[]string{HashOption}, HashSHA256},
	}
	for _, test := range tests {
		name := HashName(test.opts)
		assert.Equal(t, test.name, name)
	}
}

func TestIsSecure(t *testing.T) {
	tests := []struct {
		opts []string
		is   bool
	}{
		{nil, false},
		{[]string{}, false},
		{[]string{SecureOption}, true},
	}
	for _, test := range tests {
		is := IsSecure(test.opts)
		assert.Equal(t, test.is, is)
	}
}
