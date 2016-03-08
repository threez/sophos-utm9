// Copyright 2016 Vincent Landgraf. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package confd

// FilterObjects allows filtering of objects.
// Multiple top level filter imply "and" expression.
func (c *Conn) FilterObjects() *ObjectFilter {
	return &ObjectFilter{conn: c}
}

// ObjectFilter contains all filters
type ObjectFilter struct {
	conn             *Conn
	className        *string       // optional
	typeNames        []string      // optional
	attributeFilters []interface{} // optional
}

// Get all objects found by the filter
func (f *ObjectFilter) Get() ([]AnyObject, error) {
	var objects []AnyObject
	args := make([]interface{}, 2+len(f.attributeFilters))
	args[0] = f.className
	args[1] = f.typeNames
	for i, arg := range f.attributeFilters {
		args[i+2] = arg
	}
	err := f.conn.Request("get_objects", &objects, args...)
	return objects, err
}

// ClassName filter for passed class name (optional).
// Note: later invocations will overwrite the first name
func (f *ObjectFilter) ClassName(name string) *ObjectFilter {
	f.className = &name
	return f
}

// TypeName filter for passed type name.
// Note: Multiple invocations will add to the list of filtered types
func (f *ObjectFilter) TypeName(name string) *ObjectFilter {
	f.typeNames = append(f.typeNames, name)
	return f
}

// Eq checks if the name is equal to value
func (f *ObjectFilter) Eq(name string, value interface{}) *ObjectFilter {
	return f.filter(name, "eq", value)
}

// Ne checks if the name is not equal to value
func (f *ObjectFilter) Ne(name string, value interface{}) *ObjectFilter {
	return f.filter(name, "ne", value)
}

// Gt checks if the name is greater than value
func (f *ObjectFilter) Gt(name string, value interface{}) *ObjectFilter {
	return f.filter(name, "gt", value)
}

// Gte checks if the name is greater than or equal to value
func (f *ObjectFilter) Gte(name string, value interface{}) *ObjectFilter {
	return f.filter(name, "ge", value)
}

// Lt checks if the name is less than value
func (f *ObjectFilter) Lt(name string, value interface{}) *ObjectFilter {
	return f.filter(name, "lt", value)
}

// Lte checks if the name is less than or equal to value
func (f *ObjectFilter) Lte(name string, value interface{}) *ObjectFilter {
	return f.filter(name, "le", value)
}

// Matches checks if the name is does match (regex) to value
func (f *ObjectFilter) Matches(name string, value string) *ObjectFilter {
	return f.filter(name, "=~", value)
}

// NotMatches checks if the name is doesn't match (regex) to value
func (f *ObjectFilter) NotMatches(name string, value string) *ObjectFilter {
	return f.filter(name, "!~", value)
}

// Default checks if the name is still the default value
func (f *ObjectFilter) Default(name string) *ObjectFilter {
	f.attributeFilters = append(f.attributeFilters,
		[]interface{}{name, "default"})
	return f
}

// Or adds or (||) condition around the passed filter
func (f *ObjectFilter) Or(of *ObjectFilter) *ObjectFilter {
	return f.expression("_or", of)
}

// And adds add (&&) condition around the passed filter
func (f *ObjectFilter) And(of *ObjectFilter) *ObjectFilter {
	return f.expression("_and", of)
}

// Not adds not (!) condition around the passed filter
func (f *ObjectFilter) Not(of *ObjectFilter) *ObjectFilter {
	return f.expression("_not", of)
}

func (f *ObjectFilter) expression(exp string, of *ObjectFilter) *ObjectFilter {
	filter := make([]interface{}, len(of.attributeFilters)+1)
	filter[0] = exp
	for i, attr := range of.attributeFilters {
		filter[i+1] = attr
	}
	f.attributeFilters = append(f.attributeFilters, filter)
	return f
}

// Eq checks if the name is equal to value
func (f *ObjectFilter) filter(name string, exp string, value interface{}) *ObjectFilter {
	filter := []interface{}{name, exp, value}
	f.attributeFilters = append(f.attributeFilters, filter)
	return f
}
