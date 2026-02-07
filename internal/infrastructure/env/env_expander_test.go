package env

import (
	"os"
	"testing"
)

func TestExpander_Expand(t *testing.T) {
	expander := NewExpander()

	os.Setenv("TEST_VAR", "test_value")
	os.Setenv("HOME", "/home/user")
	defer os.Unsetenv("TEST_VAR")

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple variable", "$TEST_VAR", "test_value"},
		{"variable in text", "home is $HOME", "home is /home/user"},
		{"braced variable", "${TEST_VAR}", "test_value"},
		{"braced in text", "var=${TEST_VAR}", "var=test_value"},
		{"multiple variables", "$HOME/$TEST_VAR", "/home/user/test_value"},
		{"nonexistent variable", "$NONEXISTENT", ""},
		{"nonexistent braced", "${NONEXISTENT}", ""},
		{"no variable", "plain text", "plain text"},
		{"dollar sign only", "$", "$"},
		{"incomplete braced", "${TEST", "${TEST"},
		{"mixed", "start $TEST_VAR middle ${HOME} end", "start test_value middle /home/user end"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expander.Expand(tt.input)
			if result != tt.expected {
				t.Errorf("Expand(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestExpander_ExpandSpecialChars(t *testing.T) {
	expander := NewExpander()

	os.Setenv("VAR_WITH_UNDERSCORE", "value")
	defer os.Unsetenv("VAR_WITH_UNDERSCORE")

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"underscore in name", "$VAR_WITH_UNDERSCORE", "value"},
		{"numbers in name", "$VAR123", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expander.Expand(tt.input)
			if result != tt.expected {
				t.Errorf("Expand(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

