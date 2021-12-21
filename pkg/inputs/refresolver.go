package inputs

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

// RefResolver allows references to be resolved.
type RefResolver struct {
	schemas []*Schema
	//           k=uri     v=Schema
	pathToSchema map[string]*Schema
}

// NewRefResolver creates a reference resolver.
func NewRefResolver(schemas []*Schema) *RefResolver {
	return &RefResolver{
		schemas: schemas,
	}
}

// Init the resolver.
func (r *RefResolver) Init() error {
	r.pathToSchema = make(map[string]*Schema)
	for _, v := range r.schemas {
		if err := r.mapPaths(v); err != nil {
			return err
		}
	}
	return nil
}

// recusively generate path to schema
func getPath(schema *Schema, path string) string {
	path = schema.PathElement + "/" + path
	if schema.IsRoot() {
		return path
	}
	return getPath(schema.Parent, path)
}

// GetPath generates a path to given schema.
func (r *RefResolver) GetPath(schema *Schema) string {
	if schema.IsRoot() {
		return "#"
	}
	return getPath(schema.Parent, schema.PathElement)
}

// GetSchemaByReference returns the schema.
func (r *RefResolver) GetSchemaByReference(schema *Schema) (*Schema, error) {
	u, err := url.Parse(schema.GetRoot().ID())
	if err != nil {
		return nil, err
	}
	ref, err := url.Parse(schema.Reference)
	if err != nil {
		return nil, err
	}
	resolvedPath := u.ResolveReference(ref)
	path, ok := r.pathToSchema[resolvedPath.String()]
	if !ok {
		return nil, errors.New("refresolver.GetSchemaByReference: reference not found: " + schema.Reference)
	}
	return path, nil
}

func (r *RefResolver) mapPaths(schema *Schema) error {
	rootURI := &url.URL{}
	id := schema.ID()
	if id == "" {
		if err := r.InsertURI("#", schema); err != nil {
			return err
		}
	} else {
		var err error
		rootURI, err = url.Parse(id)
		if err != nil {
			return err
		}
		// ensure no fragment.
		rootURI.Fragment = ""
		if err := r.InsertURI(rootURI.String(), schema); err != nil {
			return err
		}
		// add as JSON pointer (?)
		if err := r.InsertURI(rootURI.String()+"#", schema); err != nil {
			return err
		}
	}
	r.updateURIs(schema, *rootURI, false, false)
	return nil
}

// create a map of base URIs
func (r *RefResolver) updateURIs(schema *Schema, baseURI url.URL, checkCurrentID bool, ignoreFragments bool) error {
	// already done for root, and if schema sets a new base URI
	if checkCurrentID {
		id := schema.ID()
		if id != "" {
			newBase, err := url.Parse(id)
			if err != nil {
				return err
			}
			// if it's a JSON fragment and we're coming from part of the tree where the baseURI has changed, we need to
			// ignore the fragment, since it won't be resolvable under the current baseURI.
			if !(strings.HasPrefix(id, "#") && ignoreFragments) {
				// map all the subschema under the new base
				resolved := baseURI.ResolveReference(newBase)
				if err := r.InsertURI(resolved.String(), schema); err != nil {
					return err
				}
				if resolved.Fragment == "" {
					if err := r.InsertURI(resolved.String()+"#", schema); err != nil {
						return err
					}
				}
				if err := r.updateURIs(schema, *resolved, false, false); err != nil {
					return err
				}
				// and continue to map all subschema under the old base (except for fragments)
				ignoreFragments = true
			}
		}
	}
	for k, subSchema := range schema.Definitions {
		newBaseURI := baseURI
		newBaseURI.Fragment += "/definitions/" + k
		if err := r.InsertURI(newBaseURI.String(), subSchema); err != nil {
			return err
		}
		r.updateURIs(subSchema, newBaseURI, true, ignoreFragments)
	}
	for k, subSchema := range schema.Properties {
		newBaseURI := baseURI
		newBaseURI.Fragment += "/properties/" + k
		if err := r.InsertURI(newBaseURI.String(), subSchema); err != nil {
			return err
		}
		r.updateURIs(subSchema, newBaseURI, true, ignoreFragments)
	}
	if schema.AdditionalProperties != nil {
		newBaseURI := baseURI
		newBaseURI.Fragment += "/additionalProperties"
		r.updateURIs((*Schema)(schema.AdditionalProperties), newBaseURI, true, ignoreFragments)
	}
	if schema.Items != nil {
		newBaseURI := baseURI
		newBaseURI.Fragment += "/items"
		r.updateURIs(schema.Items, newBaseURI, true, ignoreFragments)
	}
	return nil
}

// InsertURI to the references.
func (r *RefResolver) InsertURI(uri string, schema *Schema) error {
	if _, ok := r.pathToSchema[uri]; ok {
		return fmt.Errorf("attempted to add duplicate uri: %s/%s", schema.GetRoot().ID(), uri)
	}
	r.pathToSchema[uri] = schema
	return nil
}
