package jsonschema

import (
	"errors"
	"net/url"
	"strings"
	"fmt"
)

type RefResolver struct {
	schemas []*Schema
	//             k=docId  k=JSONPtr  v=Schema
	pathToSchema map[string]map[string]*Schema
}


func NewRefResolver(schemas []*Schema) *RefResolver {
	return &RefResolver{
		schemas: schemas,
	}
}

func (r *RefResolver) Init() error {
	r.pathToSchema = make(map[string]map[string]*Schema)

	for _, v := range r.schemas {
		if err := r.mapPaths(v); err != nil {
			return err
		}
	}

	dump := true
	if dump {
		for docId, _ := range r.pathToSchema {
			for p, v := range r.pathToSchema[docId] {
				fmt.Println(docId, p, v.TypeValue)
			}
		}
	}

	return nil
}

func (r *RefResolver) SchemaContainsPath(schema *Schema, path string) (bool, error) {
	id := schema.GetRoot().ID()
	if paths, ok := r.pathToSchema[id]; !ok {
		return false, errors.New("schema id not found in refresolver: "+id)
	} else {
		if _, ok := paths[path]; !ok {
			return false, errors.New("path not found in refresolver: "+id+path)
		} else {
			return true, nil
		}
	}
}

func getPath(schema *Schema, path string) string {
	path = schema.PathElement + "/" + path
	if schema.IsRoot() {
		return path
	} else {
		return getPath(schema.Parent, path)
	}
}

// generate a path to given schema
func (r *RefResolver) GetPath(schema *Schema) string {
	if schema.IsRoot() {
		return "#"
	} else {
		return getPath(schema.Parent, schema.PathElement)
	}

}

func (r *RefResolver) GetSchemaByReference(schema *Schema) (*Schema, error) {
	docId := schema.GetRoot().ID()

	if url, err := url.Parse(docId); err != nil {
		return nil, err
	} else {
		if ref, err := url.Parse(schema.Reference); err != nil {
			return nil, err
		} else {
			resolvedPath := url.ResolveReference(ref)
			str := resolvedPath.String()

			if path, ok := r.pathToSchema[docId][str]; !ok {
				return nil, errors.New("refresolver.GetSchemaByReference: reference not found: " + schema.Reference)
			} else {
				return path, nil
			}
		}
	}
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
func (r *RefResolver) updateURIs(schema *Schema, baseURI url.URL, checkCurrentId bool, ignoreFragments bool) error {

	// already done for root, and if schema sets a new base URI
	if checkCurrentId {
		id := schema.ID()
		if id != "" {
			if newBase, err := url.Parse(id); err != nil {
				return err
			} else {
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

func (r *RefResolver) InsertURI(uri string, schema *Schema) error {
	docId := schema.GetRoot().ID()

	if _, ok := r.pathToSchema[docId]; !ok {
		r.pathToSchema[docId] = make(map[string]*Schema)
	}

	if _, ok := r.pathToSchema[docId][uri]; ok {
		return errors.New("attempted to add duplicate uri: "+docId+"/"+uri)
	} else {
		r.pathToSchema[docId][uri] = schema
	}

	return nil
}