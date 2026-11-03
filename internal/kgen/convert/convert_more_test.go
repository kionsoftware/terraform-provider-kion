package convert

import (
	"testing"
)

func TestToSnakeCase(t *testing.T) {
	testCases := []struct {
		TestName string
		Input    string
		Expected string
	}{
		{TestName: "empty", Input: "", Expected: ""},
		{TestName: "single lower", Input: "a", Expected: "a"},
		{TestName: "single upper", Input: "A", Expected: "a"},
		{TestName: "single word lower", Input: "cheese", Expected: "cheese"},
		{TestName: "single word title", Input: "Cheese", Expected: "cheese"},
		{TestName: "two words", Input: "CityLights", Expected: "city_lights"},
		{TestName: "three words", Input: "OpenEndResource", Expected: "open_end_resource"},
		{TestName: "leading acronym two letter", Input: "OUNote", Expected: "ou_note"},
		{TestName: "leading acronym three letter", Input: "AWSAccount", Expected: "aws_account"},
		{TestName: "leading acronym IAM", Input: "IAMPolicy", Expected: "iam_policy"},
		{TestName: "two letter prefix", Input: "DBInstance", Expected: "db_instance"},
		{TestName: "multiple acronyms", Input: "DBInstanceVPCEndpoint", Expected: "db_instance_vpc_endpoint"},
		{TestName: "trailing acronym", Input: "FooBAR", Expected: "foo_bar"},
		{TestName: "arn parse", Input: "ARNParse", Expected: "arn_parse"},
		{TestName: "already snake", Input: "already_snake", Expected: "already_snake"},
		{TestName: "already snake with word", Input: "user_id", Expected: "user_id"},
		{TestName: "underscores only", Input: "___", Expected: "___"},
		{TestName: "mixed underscore camel", Input: "Foo_Bar", Expected: "foo_bar"},
		{TestName: "two caps", Input: "AB", Expected: "ab"},
		{TestName: "cap cap lower", Input: "ABc", Expected: "a_bc"},
		{TestName: "lower cap", Input: "aB", Expected: "a_b"},
		{TestName: "lower digit", Input: "a1", Expected: "a_1"},
		{TestName: "digit lower", Input: "1a", Expected: "1_a"},
		{TestName: "cap digit", Input: "A1", Expected: "a1"},
		{TestName: "acronym then digit then word", Input: "AWS3Bucket", Expected: "aws3_bucket"},
		{TestName: "word then digits", Input: "Simple123", Expected: "simple_123"},
		{TestName: "acronym word digit", Input: "HTTPServer2", Expected: "http_server_2"},
		{TestName: "single upper x", Input: "X", Expected: "x"},
		{TestName: "word digit cap word", Input: "Ab1Cd", Expected: "ab_1_cd"},
		{TestName: "acronym trailing digit", Input: "VPC2", Expected: "vpc2"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.TestName, func(t *testing.T) {
			got := ToSnakeCase(testCase.Input)
			if got != testCase.Expected {
				t.Errorf("ToSnakeCase(%q) = %q, expected %q", testCase.Input, got, testCase.Expected)
			}
		})
	}
}

// TestToHumanResNameMore adds cases beyond the existing TestToHumanResName.
func TestToHumanResNameMore(t *testing.T) {
	testCases := []struct {
		TestName string
		Input    string
		Expected string
	}{
		{TestName: "empty", Input: "", Expected: ""},
		{TestName: "single upper", Input: "A", Expected: "A"},
		{TestName: "leading acronym three letter", Input: "AWSAccount", Expected: "AWS Account"},
		{TestName: "leading acronym two letter", Input: "OUNote", Expected: "OU Note"},
		{TestName: "iam policy", Input: "IAMPolicy", Expected: "IAM Policy"},
		{TestName: "lower then acronym", Input: "abcDEF", Expected: "abc DEF"},
		{TestName: "lower cap cap lower", Input: "aBCd", Expected: "a B Cd"},
		{TestName: "acronym then camel", Input: "XMLHttpRequest", Expected: "XML Http Request"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.TestName, func(t *testing.T) {
			got := ToHumanResName(testCase.Input)
			if got != testCase.Expected {
				t.Errorf("ToHumanResName(%q) = %q, expected %q", testCase.Input, got, testCase.Expected)
			}
		})
	}
}

// TestToLowercasePrefixMore adds cases beyond the existing TestToLowercasePrefix.
func TestToLowercasePrefixMore(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want string
	}{
		{"single upper", "A", "a"},
		{"two upper", "AB", "ab"},
		{"leading digit then camel", "1Abc", "1Abc"},
		{"all digits", "123", "123"},
		{"acronym prefix lower rest", "ABCdef", "abCdef"},
		{"upper then all upper rest", "aBC", "aBC"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToLowercasePrefix(tt.s); got != tt.want {
				t.Errorf("ToLowercasePrefix(%q) = %q, want %q", tt.s, got, tt.want)
			}
		})
	}
}

func TestToProviderResourceName(t *testing.T) {
	tests := []struct {
		name      string
		first     string
		snakeName string
		want      string
	}{
		{"simple", "ignored", "my_resource", "kion_my_resource"},
		{"empty snake", "", "", "kion_"},
		{"single word", "x", "account", "kion_account"},
		{"first arg ignored", "SomethingElse", "vpc_endpoint", "kion_vpc_endpoint"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToProviderResourceName(tt.first, tt.snakeName); got != tt.want {
				t.Errorf("ToProviderResourceName(%q, %q) = %q, want %q", tt.first, tt.snakeName, got, tt.want)
			}
		})
	}
}

func TestByteHelpers(t *testing.T) {
	if !isCapitalLetter('A') || !isCapitalLetter('Z') || isCapitalLetter('a') || isCapitalLetter('0') {
		t.Errorf("isCapitalLetter classification wrong")
	}
	if !isLowercaseLetter('a') || !isLowercaseLetter('z') || isLowercaseLetter('A') || isLowercaseLetter('9') {
		t.Errorf("isLowercaseLetter classification wrong")
	}
	if !isNumeric('0') || !isNumeric('9') || isNumeric('a') || isNumeric('A') {
		t.Errorf("isNumeric classification wrong")
	}
	if got := toLowercaseLetter('A'); got != 'a' {
		t.Errorf("toLowercaseLetter('A') = %q, want %q", got, 'a')
	}
	if got := toLowercaseLetter('Z'); got != 'z' {
		t.Errorf("toLowercaseLetter('Z') = %q, want %q", got, 'z')
	}
}
