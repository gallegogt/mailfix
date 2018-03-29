package mailfix

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"regexp"
	"strings"
	"unicode"
	"unsafe"

	"github.com/smancke/mailck"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var (
	ErrInvalidUserFormat = errors.New("Invalid User Format")
	ErrInvalidHostFormat = errors.New("Invalid Host Format")
	InvalidDomain        = errors.New("Unresolvable host")

	userRegexp = regexp.MustCompile("^[a-zA-Z0-9!#$%&'*+/=?^_`{|}~.-]+$")
	hostRegexp = regexp.MustCompile("^[^\\s]+\\.[^\\s]{2,}$")
	// As per RFC 5332 secion 3.2.3: https://tools.ietf.org/html/rfc5322#section-3.2.3
	// Dots are not allowed in the beginning, end or in occurances of more than 1 in the email address
	userDotRegexp = regexp.MustCompile("(^[.]{1})|([.]{1}$)|([.]{2,})")
)

// EmailHandler
type EmailHandler interface {
	// ValidateHost valida que el host esté online
	ValidateHost() error
	// TryFixHost ententa corregir el host en caso de que no sea válido
	TryFixHost(patterns []string) (string, error)
	// ValidateFormat valida mediante Regex el formato del email
	ValidateFormat() error
}

// Email Estructura para encapsular el email y sus partes
type EmailInfo struct {
	// ToEmail texto que representa el correo completo
	ToEmail string
	// FromEmail texto que representa el correo completo
	FromEmail string
	// User parte del email que representa el usuario
	User string
	// Host parte del email que representa el HOST
	Host string
	// Normalized Email formatedao
	Normalized string
}

// TryFixHost
func (self *EmailInfo) TryFixHost() error {
	return nil
}

// ValidateHost Valida el host
func (self *EmailInfo) ValidateHost() error {
	switch self.Host {
	case "localhost", "example.com":
		return InvalidDomain
	}
	_, errMX := net.LookupMX(self.Host)
	_, errLIP := net.LookupIP(self.Host)

	if errMX != nil && errLIP != nil {
		// Only fail if both MX and A records are missing - any of the
		// two is enough for an email to be deliverable
		return InvalidDomain
	}
	return nil
}

// ValidateFormat valida mediante Regex el formato del email
func (self *EmailInfo) ValidateFormat() error {
	self.Normalized = normalizeString(self.ToEmail)

	if len(self.Normalized) < 6 || len(self.Normalized) > 254 {
		return ErrInvalidUserFormat
	}

	at := strings.LastIndex(self.Normalized, "@")
	if at <= 0 || at > len(self.Normalized)-4 {
		return ErrInvalidUserFormat
	}

	self.User = self.Normalized[:at]
	self.Host = self.Normalized[at+1:]

	if len(self.User) > 64 || userDotRegexp.MatchString(self.User) || !userRegexp.MatchString(self.User) {
		return ErrInvalidUserFormat
	}
	if !hostRegexp.MatchString(self.Host) {
		return ErrInvalidHostFormat
	}

	return nil
}

func Fix(fromEmail, toEmail string) {
	// email = normalizeString("Rocío.sylvester@gmail.com")
	eInfo := EmailInfo{
		FromEmail: fromEmail,
		ToEmail:   toEmail,
	}
	err := eInfo.ValidateFormat()
	if err == ErrInvalidUserFormat {
		return
	}

	if err := eInfo.ValidateHost(); err != nil {
		// try to fix hosts
		eInfo.TryFixHost()
	}

	result, _ := mailck.Check(fromEmail, toEmail)
	switch {
	case result.IsValid():
		fmt.Println("the mailserver accepts mails for this mailbox.")
	case result.IsError():
		fmt.Println("something went wrong in the smtp communication")
	case result.IsInvalid():
		fmt.Println(result.ResultDetail)
	}
}

// ===================================== =====================================
//			UTILS
// ===================================== =====================================
// source: http://www.golangprograms.com/data-structure-and-algorithms/golang-program-for-implementation-of-levenshtein-distance.html
// levenshteinDistance calcula la cantidad de cambios mínimos para transformar
// una cadena en otra
func levenshteinDistance(str1 []rune, str2 []rune) int {
	s1len := len(str1)
	s2len := len(str2)
	column := make([]int, len(str1)+1)

	for y := 1; y <= s1len; y++ {
		column[y] = y
	}
	for x := 1; x <= s2len; x++ {
		column[0] = x
		lastkey := x - 1
		for y := 1; y <= s1len; y++ {
			oldkey := column[y]
			var incr int
			if str1[y-1] != str2[x-1] {
				incr = 1
			}

			column[y] = min(column[y]+1, min(column[y-1]+1, lastkey+incr))
			lastkey = oldkey
		}
	}
	return column[s1len]
}

func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

// NormalizeString remueve o normaliza los caracteres no ASCII
func normalizeString(str string) string {
	isMn := func(r rune) bool {
		return unicode.Is(unicode.Mn, r) // Mn: nonspacing marks
	}
	t := transform.Chain(norm.NFD, transform.RemoveFunc(isMn), norm.NFC)
	r := transform.NewReader(strings.NewReader(str), t)
	normalized := stringFromReader(r)
	// trim spaces
	normalized = strings.TrimSpace(normalized)
	// remove end dot char
	normalized = strings.TrimRight(normalized, ".")
	// remove left dot char
	normalized = strings.TrimLeft(normalized, ".")
	// Lower Cases
	normalized = strings.ToLower(normalized)
	return normalized
}

// stringFromReader convierte de io.Reader a String
func stringFromReader(reader io.Reader) string {
	buf := new(bytes.Buffer)
	buf.ReadFrom(reader)
	b := buf.Bytes()
	str := *(*string)(unsafe.Pointer(&b))
	return str
}
