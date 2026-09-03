package variable

import (
	"fmt"
	"regexp"
	"strings"

	"highway/storage"
)

var variablePattern = regexp.MustCompile(`\{\{\s*([^{}]+?)\s*\}\}`)

func ExpandRequestSnapshot(snapshot storage.RequestSnapshot, variables [][2]string) (storage.RequestSnapshot, error) {
	values, err := VariableValues(variables)
	if err != nil {
		return storage.RequestSnapshot{}, err
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

	if snapshot.URL, err = expand(snapshot.URL); err != nil {
		return storage.RequestSnapshot{}, err
	}
	if snapshot.RawBody, err = expand(snapshot.RawBody); err != nil {
		return storage.RequestSnapshot{}, err
	}
	if snapshot.QueryParams, err = ExpandPairs(snapshot.QueryParams, expand); err != nil {
		return storage.RequestSnapshot{}, err
	}
	if snapshot.Headers, err = ExpandPairs(snapshot.Headers, expand); err != nil {
		return storage.RequestSnapshot{}, err
	}
	if snapshot.Form, err = ExpandPairs(snapshot.Form, expand); err != nil {
		return storage.RequestSnapshot{}, err
	}
	if snapshot.Multipart, err = ExpandPairs(snapshot.Multipart, expand); err != nil {
		return storage.RequestSnapshot{}, err
	}

	if snapshot.Auth.Token, err = expand(snapshot.Auth.Token); err != nil {
		return storage.RequestSnapshot{}, err
	}
	if snapshot.Auth.BasicUser, err = expand(snapshot.Auth.BasicUser); err != nil {
		return storage.RequestSnapshot{}, err
	}
	if snapshot.Auth.BasicPass, err = expand(snapshot.Auth.BasicPass); err != nil {
		return storage.RequestSnapshot{}, err
	}
	if snapshot.Auth.APIKeyName, err = expand(snapshot.Auth.APIKeyName); err != nil {
		return storage.RequestSnapshot{}, err
	}
	if snapshot.Auth.APIKeyValue, err = expand(snapshot.Auth.APIKeyValue); err != nil {
		return storage.RequestSnapshot{}, err
	}
	if snapshot.Auth.OAuth1ConsumerKey, err = expand(snapshot.Auth.OAuth1ConsumerKey); err != nil {
		return storage.RequestSnapshot{}, err
	}
	if snapshot.Auth.OAuth1ConsumerSecret, err = expand(snapshot.Auth.OAuth1ConsumerSecret); err != nil {
		return storage.RequestSnapshot{}, err
	}
	if snapshot.Auth.OAuth1AccessToken, err = expand(snapshot.Auth.OAuth1AccessToken); err != nil {
		return storage.RequestSnapshot{}, err
	}
	if snapshot.Auth.OAuth1TokenSecret, err = expand(snapshot.Auth.OAuth1TokenSecret); err != nil {
		return storage.RequestSnapshot{}, err
	}
	if snapshot.Auth.TokenURL, err = expand(snapshot.Auth.TokenURL); err != nil {
		return storage.RequestSnapshot{}, err
	}
	if snapshot.Auth.ClientID, err = expand(snapshot.Auth.ClientID); err != nil {
		return storage.RequestSnapshot{}, err
	}
	if snapshot.Auth.ClientSecret, err = expand(snapshot.Auth.ClientSecret); err != nil {
		return storage.RequestSnapshot{}, err
	}
	if snapshot.Auth.Username, err = expand(snapshot.Auth.Username); err != nil {
		return storage.RequestSnapshot{}, err
	}
	if snapshot.Auth.Password, err = expand(snapshot.Auth.Password); err != nil {
		return storage.RequestSnapshot{}, err
	}
	if snapshot.Auth.Scope, err = expand(snapshot.Auth.Scope); err != nil {
		return storage.RequestSnapshot{}, err
	}

	return snapshot, nil
}

func VariableValues(variables [][2]string) (map[string]string, error) {
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

func ExpandPairs(pairs [][2]string, expand func(string) (string, error)) ([][2]string, error) {
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
