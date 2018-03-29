package mailfix

import "testing"

// ----------------------------------------------------------------------------
var LDistanceTests = []struct {
	textA    []rune // Texto A
	textB    []rune // Texto B
	expected int    // expected result
}{
	{[]rune("r0.syl@gmail.com"), []rune("r0.syl@gmail.com"), 0},
	{[]rune("hello"), []rune("elloh"), 2},
	{[]rune("hello"), []rune("ello"), 1},
	{[]rune("hello"), []rune("olleh"), 4},
	{[]rune("sitting"), []rune("kitten"), 3},
	{[]rune("gmail.com"), []rune("gmai.com"), 1},
	{[]rune("gmail.com"), []rune("gmail.cm"), 1},
	{[]rune("hotmail.com"), []rune("hotmail.con"), 1},
}

func TestLDistance(t *testing.T) {
	for _, tt := range LDistanceTests {
		actual := levenshteinDistance(tt.textA, tt.textB)
		if actual != tt.expected {
			t.Errorf("levenshteinDistance('%s', '%s'): expected '%d', actual '%d'", string(tt.textA), string(tt.textB), tt.expected, actual)
		}
	}
}

// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------

var EmailFormatTests = []struct {
	email    string // input email string
	expected error  // expected result
}{
	{"r0.syl@gmail.com", nil},
	{".syl@gmail.com", nil},
	{"@gmail", ErrInvalidUserFormat},
	{"email@gmail", ErrInvalidHostFormat},
	{"noemail@gmail.c", ErrInvalidHostFormat},
}

func TestEmailFormat(t *testing.T) {
	for _, tt := range EmailFormatTests {
		eInfo := EmailInfo{ToEmail: tt.email}
		actual := eInfo.ValidateFormat()
		if actual != tt.expected {
			t.Errorf("EmailInfo.ValidateFormat(): expected '%s', actual '%s'", tt.expected, actual)
		}
	}
}

// ----------------------------------------------------------------------------
var EmailStringTests = []struct {
	email    string // input email string
	expected string // expected result
}{
	{"Rocío.sylvester@gmail.com", "rocio.sylvester@gmail.com"},
	{"sylvester@gmail.com.", "sylvester@gmail.com"},
	{"UPPER_EMAIL_IS_INVALID@gmail.com", "upper_email_is_invalid@gmail.com"},
	{".dontstartwithdot@gmail.com", "dontstartwithdot@gmail.com"},
	{"Ñooo@gmail.com", "nooo@gmail.com"},
	{"Ñoño@gmail.com", "nono@gmail.com"},
}

func TestNormalizeString(t *testing.T) {
	for _, tt := range EmailStringTests {
		actual := normalizeString(tt.email)
		if actual != tt.expected {
			t.Errorf("NormalizeString(%s): expected %s, actual %s", tt.email, tt.expected, actual)
		}
	}
}

// ----------------------------------------------------------------------------
