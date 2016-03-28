package hstspreload

import (
	"fmt"
	"testing"
)

func ExampleCheckDomain() {
	issues := CheckDomain("wikipedia.org")
	fmt.Printf("%v", issues)
}

/******** Utility functions tests. ********/

func TestCheckDomainFormat(t *testing.T) {
	expectIssuesEqual(t, checkDomainFormat(".example.com"),
		NewIssues().addErrorf("Domain name error: begins with `.`"))
	expectIssuesEqual(t, checkDomainFormat("example.com."),
		NewIssues().addErrorf("Domain name error: ends with `.`"))
	expectIssuesEqual(t, checkDomainFormat("example..com"),
		NewIssues().addErrorf("Domain name error: contains `..`"))
	expectIssuesEqual(t, checkDomainFormat("example"),
		NewIssues().addErrorf("Domain name error: must have at least two labels."))
	expectIssuesEqual(t, checkDomainFormat("example&co.com"),
		NewIssues().addErrorf("Domain name error: contains invalid characters."))
}

func TestCheckEffectiveTLDPlusOne(t *testing.T) {
	expectIssuesEqual(t, checkEffectiveTLDPlusOne("subdomain.example.com"),
		NewIssues().addErrorf("Domain error: `subdomain.example.com` is not eTLD+1. Please preload `example.com` instead."))
}

/******** Real domain tests. ********/

// Avoid hitting the network for short tests.
// This gives us performant, deterministic, and offline testing.
func skipIfShort(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping domain test.")
	}
}

func TestCheckDomainIncompleteChain(t *testing.T) {
	skipIfShort(t)
	expectIssuesEqual(t, CheckDomain("incomplete-chain.badssl.com"),
		Issues{
			Errors: []string{
				"Domain error: `incomplete-chain.badssl.com` is not eTLD+1. Please preload `badssl.com` instead.",
				"Cannot connect using TLS (\"Get https://incomplete-chain.badssl.com: x509: certificate signed by unknown authority\"). This might be caused by an incomplete certificate chain, which causes issues on mobile devices. Check out your site at https://www.ssllabs.com/ssltest/",
			},
			Warnings: []string{},
		},
	)
}

func TestCheckDomainSHA1(t *testing.T) {
	skipIfShort(t)
	expectIssuesEqual(t, CheckDomain("sha1.badssl.com"),
		Issues{
			Errors: []string{
				"Domain error: `sha1.badssl.com` is not eTLD+1. Please preload `badssl.com` instead.",
				"One or more of the certificates in your certificate chain is signed with SHA-1. This needs to be replaced. See https://security.googleblog.com/2015/12/an-update-on-sha-1-certificates-in.html. (The first SHA-1 certificate found has a common-name of \"*.badssl.com\".)",
				"Response error: No HSTS headers are present on the response.",
			},
			Warnings: []string{},
		},
	)
}

func TestCheckDomainWithValidHSTS(t *testing.T) {
	skipIfShort(t)
	expectIssuesEmpty(t, CheckDomain("wikipedia.org"))
}

func TestCheckDomainSubdomain(t *testing.T) {
	skipIfShort(t)
	expectIssuesEqual(t, CheckDomain("en.wikipedia.org"),
		NewIssues().addErrorf("Domain error: `en.wikipedia.org` is not eTLD+1. Please preload `wikipedia.org` instead."),
	)
}

func TestCheckDomainWithoutHSTS(t *testing.T) {
	skipIfShort(t)
	expectIssuesEqual(t, CheckDomain("example.com"),
		NewIssues().addErrorf("Response error: No HSTS headers are present on the response."))
}

func TestCheckDomainBogusDomain(t *testing.T) {
	skipIfShort(t)
	expectIssuesEqual(t, CheckDomain("example.notadomain"),
		NewIssues().addErrorf(`Cannot connect using TLS ("Get https://example.notadomain: dial tcp: lookup example.notadomain: no such host"). This might be caused by an incomplete certificate chain, which causes issues on mobile devices. Check out your site at https://www.ssllabs.com/ssltest/`))
}
