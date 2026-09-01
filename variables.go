package main

import (
	"fmt"
	"regexp"
	"strings"
)

var variablePattern = regexp.MustCompile(`\{\{\s*([^{}]+?)\s*\}\}`)

func expandRequestSnapshot(snapshot requestSnapshot, variables [][2]string) (requestSnapshot, error) {
	values, err := variableValues(variables)
	if err != nil {
		return requestSnapshot{}, err
	}

	expand := func(value string) (string, error) {
		var missing string
		result := variablePattern.ReplaceAllStringFunc(value, func(match string) string {
			name := strings.TrimSpace(variablePattern.FindStringSubmatch(match)[1])
			resolved, exists := values[name]
			if !exists {
				missing = name
				return match
			}
			return resolved
		})
		if missing != "" {
			return "", fmt.Errorf("variável não encontrada: %s", missing)
		}
		return result, nil
	}

	if snapshot.url, err = expand(snapshot.url); err != nil {
		return requestSnapshot{}, err
	}
	if snapshot.rawBody, err = expand(snapshot.rawBody); err != nil {
		return requestSnapshot{}, err
	}
	if snapshot.queryParams, err = expandPairs(snapshot.queryParams, expand); err != nil {
		return requestSnapshot{}, err
	}
	if snapshot.headers, err = expandPairs(snapshot.headers, expand); err != nil {
		return requestSnapshot{}, err
	}
	if snapshot.form, err = expandPairs(snapshot.form, expand); err != nil {
		return requestSnapshot{}, err
	}
	if snapshot.multipart, err = expandPairs(snapshot.multipart, expand); err != nil {
		return requestSnapshot{}, err
	}

	if snapshot.auth.Token, err = expand(snapshot.auth.Token); err != nil {
		return requestSnapshot{}, err
	}
	if snapshot.auth.BasicUser, err = expand(snapshot.auth.BasicUser); err != nil {
		return requestSnapshot{}, err
	}
	if snapshot.auth.BasicPass, err = expand(snapshot.auth.BasicPass); err != nil {
		return requestSnapshot{}, err
	}
	if snapshot.auth.APIKeyName, err = expand(snapshot.auth.APIKeyName); err != nil {
		return requestSnapshot{}, err
	}
	if snapshot.auth.APIKeyValue, err = expand(snapshot.auth.APIKeyValue); err != nil {
		return requestSnapshot{}, err
	}
	if snapshot.auth.OAuth1ConsumerKey, err = expand(snapshot.auth.OAuth1ConsumerKey); err != nil {
		return requestSnapshot{}, err
	}
	if snapshot.auth.OAuth1ConsumerSecret, err = expand(snapshot.auth.OAuth1ConsumerSecret); err != nil {
		return requestSnapshot{}, err
	}
	if snapshot.auth.OAuth1AccessToken, err = expand(snapshot.auth.OAuth1AccessToken); err != nil {
		return requestSnapshot{}, err
	}
	if snapshot.auth.OAuth1TokenSecret, err = expand(snapshot.auth.OAuth1TokenSecret); err != nil {
		return requestSnapshot{}, err
	}
	if snapshot.auth.TokenURL, err = expand(snapshot.auth.TokenURL); err != nil {
		return requestSnapshot{}, err
	}
	if snapshot.auth.ClientID, err = expand(snapshot.auth.ClientID); err != nil {
		return requestSnapshot{}, err
	}
	if snapshot.auth.ClientSecret, err = expand(snapshot.auth.ClientSecret); err != nil {
		return requestSnapshot{}, err
	}
	if snapshot.auth.Username, err = expand(snapshot.auth.Username); err != nil {
		return requestSnapshot{}, err
	}
	if snapshot.auth.Password, err = expand(snapshot.auth.Password); err != nil {
		return requestSnapshot{}, err
	}
	if snapshot.auth.Scope, err = expand(snapshot.auth.Scope); err != nil {
		return requestSnapshot{}, err
	}

	return snapshot, nil
}

func variableValues(variables [][2]string) (map[string]string, error) {
	values := make(map[string]string, len(variables))
	for _, variable := range variables {
		name := strings.TrimSpace(variable[0])
		if name == "" {
			continue
		}
		if _, exists := values[name]; exists {
			return nil, fmt.Errorf("a variável %q está definida mais de uma vez", name)
		}
		values[name] = variable[1]
	}
	return values, nil
}

func expandPairs(pairs [][2]string, expand func(string) (string, error)) ([][2]string, error) {
	result := make([][2]string, len(pairs))
	for i, pair := range pairs {
		key, err := expand(pair[0])
		if err != nil {
			return nil, err
		}
		value, err := expand(pair[1])
		if err != nil {
			return nil, err
		}
		result[i] = [2]string{key, value}
	}
	return result, nil
}
