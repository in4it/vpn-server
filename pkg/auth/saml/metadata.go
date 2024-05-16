package saml

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/russellhaering/gosaml2/types"
)

func (s *saml) HasValidMetadataURL(metadataURL string) (bool, error) {
	metadataURLParsed, err := url.Parse(metadataURL)
	if err != nil {
		return false, fmt.Errorf("url parse error: %s", err)
	}
	_, err = getMetadata(metadataURLParsed.String())
	if err != nil {
		return false, fmt.Errorf("fetch metadata error: %s", err)
	}
	return true, nil
}

func getMetadata(metadataURL string) (types.EntityDescriptor, error) {
	metadata := types.EntityDescriptor{}

	res, err := http.Get(metadataURL)
	if err != nil {
		return metadata, fmt.Errorf("can't retrieve saml metadata: %s", err)
	}

	rawMetadata, err := io.ReadAll(res.Body)
	if err != nil {
		return metadata, fmt.Errorf("can't read saml cert data: %s", err)
	}

	err = xml.Unmarshal(rawMetadata, &metadata)
	if err != nil {
		return metadata, fmt.Errorf("can't decode saml cert data: %s", err)
	}
	return metadata, nil
}
