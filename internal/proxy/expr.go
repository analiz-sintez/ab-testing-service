package proxy

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/expr-lang/expr"
)

// evaluateExpression evaluates an expression using the expr-lang/expr library
// and returns the result and any error
func evaluateExpression(expression string, env map[string]interface{}) (interface{}, error) {
	program, err := expr.Compile(expression, expr.Env(env))
	if err != nil {
		return nil, fmt.Errorf("error compiling expression: %w", err)
	}

	result, err := expr.Run(program, env)
	if err != nil {
		return nil, fmt.Errorf("error evaluating expression: %w", err)
	}

	return result, nil
}

// evaluateBooleanExpression evaluates an expression and returns true if it evaluates to a boolean true,
// false otherwise
func evaluateBooleanExpression(expression string, env map[string]interface{}) (bool, error) {
	result, err := evaluateExpression(expression, env)
	if err != nil {
		return false, err
	}

	boolResult, ok := result.(bool)
	if !ok {
		return false, fmt.Errorf("expression did not return a boolean, got %T: %v", result, result)
	}

	return boolResult, nil
}

// evaluateStringExpression evaluates an expression and returns the string result,
// or an error if the result is not a string
func evaluateStringExpression(expression string, env map[string]interface{}) (string, error) {
	result, err := evaluateExpression(expression, env)
	if err != nil {
		return "", err
	}

	stringResult, ok := result.(string)
	if !ok {
		return "", fmt.Errorf("expression did not return a string, got %T: %v", result, result)
	}

	return stringResult, nil
}

// getTargetByExpr evaluates an expression using expr-lang/expr and returns the matching target
//
// Expressions can be used in two ways:
//
//  1. Single expression in the Expr field that returns a target ID:
//     "headers['user-agent'] contains 'iPhone' ? 'target-1' : 'target-2'"
//
//  2. Multiple expressions in the Values map, each evaluating to a boolean:
//     {
//     "target-1": "headers['user-agent'] contains 'iPhone'",
//     "target-2": "query['version'] == '2'"
//     }
//
// Available variables in expressions:
// - headers: map of request headers (string keys, string values)
// - query: map of query parameters (string keys, string values)
// - cookies: map of cookies (string keys, string values)
// - method: request method (GET, POST, etc.)
// - path: request path
// - host: request host
// - ip: client IP address
// - referer: request referer
// - protocol: request protocol
// - url: full request URL
//
// For more information on the expression syntax, see: https://github.com/expr-lang/expr
func (p *Proxy) getTargetByExpr(r *http.Request) *Target {
	// Create environment with request data for the expression
	env := createExpressionEnv(r)

	// If there's a specific expression in the Expr field, evaluate it
	if p.Config.Condition.Expr != "" {
		targetID, err := evaluateStringExpression(p.Config.Condition.Expr, env)
		if err != nil {
			log.Printf("Error evaluating expression for proxy %s: %v", p.ID, err)
			return p.getTargetById(p.Config.Condition.Default)
		}

		if target := p.getTargetById(targetID); target != nil {
			return target
		}
		log.Printf("Expression returned target ID %s, but no matching target found for proxy %s", targetID, p.ID)
	} else {
		// Evaluate each expression in Values map
		for targetID, expression := range p.Config.Condition.Values {
			isMatch, err := evaluateBooleanExpression(expression, env)
			if err != nil {
				log.Printf("Error evaluating expression for target %s in proxy %s: %v", targetID, p.ID, err)
				continue
			}

			if isMatch {
				if target := p.getTargetById(targetID); target != nil {
					return target
				}
				log.Printf("Expression evaluated to true for target ID %s, but no matching target found for proxy %s", targetID, p.ID)
			}
		}
		log.Printf("No expressions evaluated to true for proxy %s", p.ID)
	}

	// Fall back to default target
	defaultTarget := p.getTargetById(p.Config.Condition.Default)
	if defaultTarget == nil {
		log.Printf("Default target %s not found for proxy %s", p.Config.Condition.Default, p.ID)
	}
	return defaultTarget
}

// Helper functions to convert request data to maps for expressions
func headersToMap(headers http.Header) map[string]string {
	result := make(map[string]string)
	for name, values := range headers {
		if len(values) > 0 {
			result[strings.ToLower(name)] = values[0]
		}
	}
	return result
}

func queryToMap(query url.Values) map[string]string {
	result := make(map[string]string)
	for name, values := range query {
		if len(values) > 0 {
			result[name] = values[0]
		}
	}
	return result
}

func cookiesToMap(r *http.Request) map[string]string {
	result := make(map[string]string)
	for _, cookie := range r.Cookies() {
		result[cookie.Name] = cookie.Value
	}
	return result
}

// createExpressionEnv creates an environment map for expression evaluation
// with all available request data
//
// Example expressions:
//
//  1. Route based on user agent:
//     "headers['user-agent'] contains 'iPhone' ? 'mobile-target' : 'desktop-target'"
//
//  2. Route based on query parameter:
//     "query['version'] == '2' ? 'new-target' : 'old-target'"
//
//  3. Route based on cookie:
//     "cookies['user_type'] == 'premium' ? 'premium-target' : 'free-target'"
//
//  4. Route based on country (using X-Country header):
//     "headers['x-country'] in ['US', 'CA'] ? 'na-target' : 'global-target'"
//
//  5. Route based on multiple conditions:
//     "headers['user-agent'] contains 'iPhone' && query['version'] == '2' ? 'iphone-v2-target' : 'default-target'"
//
//  6. Route based on path:
//     "path startsWith '/api' ? 'api-target' : 'web-target'"
//
//  7. Route based on IP address:
//     "ip startsWith '192.168.' ? 'internal-target' : 'external-target'"
//
//  8. A/B testing with percentage split (using a hash of the IP):
//     "len(ip) % 2 == 0 ? 'a-target' : 'b-target'"
//
//  9. Random traffic steering with percentage split:
//     "random(1, 100) <= 70 ? 'a-target' : 'b-target'"
//
//  10. Weighted random traffic steering:
//     "random(1, 100) <= 70 ? 'a-target' : (random(1, 100) <= 50 ? 'b-target' : 'c-target')"
//
//  11. Consistent random traffic steering (same user always gets same target):
//     "randomUser(1, 100) <= 70 ? 'a-target' : 'b-target'"
//
//  12. Cookie-based consistent traffic steering:
//     "randomCookie('ab_test', 1, 100) <= 70 ? 'a-target' : 'b-target'"
//
//  13. Parameter-based consistent traffic steering:
//     "randomParam(query['user_id'], 1, 100) <= 70 ? 'a-target' : 'b-target'"
func createExpressionEnv(r *http.Request) map[string]interface{} {
	// Seed the random number generator with the current time
	rand.Seed(time.Now().UnixNano())

	// Get client IP for user-based random function
	clientIP := getClientIP(r)

	// Create a hash of the IP address for consistent random routing
	ipHash := int64(0)
	for _, c := range clientIP {
		ipHash = ipHash*31 + int64(c)
	}

	// Create a separate random source for user-based randomization
	userRand := rand.New(rand.NewSource(ipHash))

	// Get cookies map for cookie-based randomization
	cookiesMap := cookiesToMap(r)

	// Get query parameters map
	queryMap := queryToMap(r.URL.Query())

	// Get headers map
	headersMap := headersToMap(r.Header)

	return map[string]interface{}{
		"headers":  headersMap,
		"query":    queryMap,
		"cookies":  cookiesMap,
		"method":   r.Method,
		"path":     r.URL.Path,
		"host":     r.Host,
		"ip":       clientIP,
		"referer":  r.Referer(),
		"protocol": r.Proto,
		"url":      r.URL.String(),
		// Add random function for traffic steering (different result each time)
		"random": func(min, max int) int {
			if min > max {
				min, max = max, min // Swap if min > max
			}
			return min + rand.Intn(max-min+1)
		},
		// Add user-based random function for consistent traffic steering
		// (same user always gets same result)
		"randomUser": func(min, max int) int {
			if min > max {
				min, max = max, min // Swap if min > max
			}
			return min + userRand.Intn(max-min+1)
		},
		// Add cookie-based random function for consistent traffic steering
		// (same cookie value always gets same result)
		"randomCookie": func(cookieName string, min, max int) int {
			if min > max {
				min, max = max, min // Swap if min > max
			}

			// If cookie exists, use its value to seed the random generator
			if cookieValue, exists := cookiesMap[cookieName]; exists && cookieValue != "" {
				// Create a hash of the cookie value
				cookieHash := int64(0)
				for _, c := range cookieValue {
					cookieHash = cookieHash*31 + int64(c)
				}

				// Create a random source based on the cookie hash
				cookieRand := rand.New(rand.NewSource(cookieHash))
				return min + cookieRand.Intn(max-min+1)
			}

			// If cookie doesn't exist, fall back to user-based randomization
			return min + userRand.Intn(max-min+1)
		},
		// Add parameter-based random function for consistent traffic steering
		// (same parameter value always gets same result)
		"randomParam": func(paramValue string, min, max int) int {
			if min > max {
				min, max = max, min // Swap if min > max
			}

			if paramValue != "" {
				// Create a hash of the parameter value
				paramHash := int64(0)
				for _, c := range paramValue {
					paramHash = paramHash*31 + int64(c)
				}

				// Create a random source based on the parameter hash
				paramRand := rand.New(rand.NewSource(paramHash))
				return min + paramRand.Intn(max-min+1)
			}

			// If parameter doesn't exist, fall back to user-based randomization
			return min + userRand.Intn(max-min+1)
		},
	}
}
