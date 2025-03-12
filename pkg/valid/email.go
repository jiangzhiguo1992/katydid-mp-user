package valid

import (
	"net"
	"regexp"
	"strings"
)

// EmailComponents holds the parsed parts of an email address
type EmailComponents struct {
	Address  string // Full email address
	Username string // Local part before @
	Domain   string // Domain part after @
	TLD      string // Top-level domain (last part of domain)
}

// IsEmail validates each component of an email address
func IsEmail(email string) (*EmailComponents, bool) {
	// First parse the email
	components, ok := parseEmail(email)
	if !ok {
		return nil, false
	}

	// Username validation
	if !IsEmailUsername(components.Username) {
		return components, false
	}

	// Domain validation
	if !IsEmailDomain(components.Domain) {
		return components, false
	}

	return components, true
}

// IsEmailUsername validates the local part of an email
func IsEmailUsername(username string) bool {
	username = strings.TrimSpace(username)
	if username == "" {
		return false
	} else if len(username) > 64 {
		// RFC 5321 local part max length
		return false
	}

	basicRegex := regexp.MustCompile(`^[a-zA-Z0-9!#$%&'*+\-/=?^_\x60{|}~.]+$`)
	return basicRegex.MatchString(username)
}

// IsEmailDomain validates the domain part of an email
func IsEmailDomain(domain string) bool {
	domain = strings.TrimSpace(domain)
	if domain == "" {
		return false
	} else if len(domain) > 255 {
		return false
	}

	// Check for MX records
	if _, err := net.LookupMX(domain); err != nil {
		return false
	}

	// Basic domain validation
	domainRegex := regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9\-]+(\.[a-zA-Z0-9][a-zA-Z0-9\-]+)*\.[a-zA-Z]{2,}$`)
	return domainRegex.MatchString(domain)
}

// parseEmail splits an email into its components
func parseEmail(email string) (*EmailComponents, bool) {
	// Remove whitespace
	email = strings.TrimSpace(email)

	// Convert to lowercase (though technically username can be case-sensitive)
	email = strings.ToLower(email)

	// Check if email is empty
	if email == "" {
		return nil, false
	}

	// Split by @ symbol
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return nil, false
	}

	username := parts[0]
	domain := parts[1]

	// Extract TLD
	domainParts := strings.Split(domain, ".")
	if len(domainParts) < 2 {
		return nil, false
	}

	tld := domainParts[len(domainParts)-1]

	return &EmailComponents{
		Address:  email,
		Username: username,
		Domain:   domain,
		TLD:      tld,
	}, true
}

// FindEmailsInText extracts and counts valid email addresses in a text
func FindEmailsInText(text string) []string {
	emailRegex := regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`)
	matches := emailRegex.FindAllString(text, -1)

	validEmails := make([]string, 0)
	for _, match := range matches {
		if _, ok := IsEmail(match); ok {
			validEmails = append(validEmails, match)
		}
	}

	return validEmails
}
