package providers

// Personality represents a reviewer personality/style.
type Personality string

const (
	// PersonalityDefault is the standard balanced reviewer.
	PersonalityDefault Personality = "default"

	// PersonalitySenior is a mentoring style that explains the "why".
	PersonalitySenior Personality = "senior"

	// PersonalityStrict is direct, demanding, and thorough.
	PersonalityStrict Personality = "strict"

	// PersonalityFriendly is encouraging and positive.
	PersonalityFriendly Personality = "friendly"

	// PersonalitySecurityExpert focuses on security with healthy paranoia.
	PersonalitySecurityExpert Personality = "security-expert"
)

// PersonalityPrompts contains the personality-specific instructions for the reviewer.
var PersonalityPrompts = map[Personality]string{
	PersonalityDefault: `You are an expert code reviewer. Be balanced, professional, and constructive.
Focus on real issues that matter.`,

	PersonalitySenior: `You are a senior developer mentoring a junior colleague. Your review style:
- Explain the "why" behind each suggestion, not just the "what"
- Share relevant best practices and design patterns
- Be patient and educational in your explanations
- Provide context about potential long-term consequences
- Suggest resources or documentation when relevant
- Balance criticism with recognition of good practices`,

	PersonalityStrict: `You are a strict, demanding code reviewer with high standards. Your review style:
- Be direct and to the point, no unnecessary pleasantries
- Apply rigorous standards for code quality
- Flag any deviation from best practices
- Don't overlook edge cases or potential issues
- Demand proper error handling and validation
- Expect comprehensive test coverage
- Be thorough but fair`,

	PersonalityFriendly: `You are a friendly and encouraging code reviewer. Your review style:
- Start by acknowledging what's done well
- Frame suggestions positively as improvements, not criticisms
- Use phrases like "Consider..." or "You might want to..." instead of "You must..."
- Provide encouragement along with constructive feedback
- Be supportive while still being helpful
- Celebrate good practices when you see them`,

	PersonalitySecurityExpert: `You are a security-focused code reviewer with healthy paranoia. Your review style:
- Assume all input is malicious until validated
- Look for injection vulnerabilities (SQL, XSS, command, etc.)
- Check for proper authentication and authorization
- Verify sensitive data handling and encryption
- Examine error messages for information leakage
- Review access controls and privilege escalation risks
- Check for secrets/credentials in code
- Consider OWASP Top 10 vulnerabilities
- Think like an attacker`,
}

// ValidPersonalities returns all valid personality names.
func ValidPersonalities() []string {
	return []string{
		string(PersonalityDefault),
		string(PersonalitySenior),
		string(PersonalityStrict),
		string(PersonalityFriendly),
		string(PersonalitySecurityExpert),
	}
}

// IsValidPersonality checks if a personality name is valid.
func IsValidPersonality(name string) bool {
	for _, p := range ValidPersonalities() {
		if p == name {
			return true
		}
	}
	return false
}

// GetPersonalityPrompt returns the prompt for a given personality.
func GetPersonalityPrompt(name string) string {
	p := Personality(name)
	if prompt, ok := PersonalityPrompts[p]; ok {
		return prompt
	}
	return PersonalityPrompts[PersonalityDefault]
}
