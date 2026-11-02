package agentworkspace

import (
	"reflect"
	"strings"
)

var allowedCanvasNodeTypes = map[string]struct{}{
	"script": {}, "text": {}, "image": {}, "video": {}, "audio": {}, "director": {},
}

func ValidateCanvasForStorage(canvas Canvas) error {
	for _, node := range canvas.Nodes {
		if _, ok := allowedCanvasNodeTypes[node.Type]; !ok {
			return ErrCanvasInvalid
		}
		if err := validateCanvasMetadata(node.AssetRef); err != nil {
			return err
		}
		if err := validateCanvasMetadata(node.Config); err != nil {
			return err
		}
	}
	return nil
}

func validateCanvasMetadata(values map[string]any) error {
	for key, value := range values {
		if forbiddenCanvasMetadataKey(key) {
			return ErrCanvasInvalid
		}
		if err := validateCanvasMetadataValue(reflect.ValueOf(value)); err != nil {
			return err
		}
	}
	return nil
}

func validateCanvasMetadataValue(value reflect.Value) error {
	if !value.IsValid() {
		return nil
	}
	if value.Type() == reflect.TypeOf([]byte(nil)) {
		return ErrCanvasInvalid
	}
	switch value.Kind() {
	case reflect.Interface, reflect.Pointer:
		if value.IsNil() {
			return nil
		}
		return validateCanvasMetadataValue(value.Elem())
	case reflect.Map:
		if value.Type().Key().Kind() != reflect.String {
			return ErrCanvasInvalid
		}
		iter := value.MapRange()
		for iter.Next() {
			if forbiddenCanvasMetadataKey(iter.Key().String()) {
				return ErrCanvasInvalid
			}
			if err := validateCanvasMetadataValue(iter.Value()); err != nil {
				return err
			}
		}
	case reflect.Slice, reflect.Array:
		for index := 0; index < value.Len(); index++ {
			if err := validateCanvasMetadataValue(value.Index(index)); err != nil {
				return err
			}
		}
	}
	return nil
}

func forbiddenCanvasMetadataKey(key string) bool {
	var normalized strings.Builder
	for _, char := range strings.ToLower(key) {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') {
			normalized.WriteRune(char)
		}
	}
	value := normalized.String()
	if value == "objectkey" || value == "storagekey" || value == "mediabytes" || value == "mediadata" || value == "base64" || value == "binary" || value == "blob" || value == "bytes" {
		return true
	}
	return strings.Contains(value, "token") || strings.Contains(value, "secret") || strings.Contains(value, "password") || strings.Contains(value, "authorization") || strings.Contains(value, "credential") || strings.Contains(value, "apikey") || strings.Contains(value, "accesskey") || strings.Contains(value, "privatekey")
}
