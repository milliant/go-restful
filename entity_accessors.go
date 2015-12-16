package restful

// Copyright 2015 Ernest Micklei. All rights reserved.
// Use of this source code is governed by a license
// that can be found in the LICENSE file.

import (
	"encoding/json"
	"encoding/xml"
	"strings"
	"sync"
)

// EntityReaderWriter can read and write values using an encoding such as JSON,XML.
type EntityReaderWriter interface {
	// Read a serialized version of the value from the request.
	// The Request may have a decompressing reader. Depends on Content-Encoding.
	Read(req *Request, v interface{}) error

	// Write a serialized version of the value on the response.
	// The Response may have a compressing writer. Depends on Accept-Encoding.
	// status should be a valid Http Status code
	Write(resp *Response, status int, v interface{}) error
}

// entityAccessRegistry is a singleton   entityAccessRegistry是一个单例
//entityReaderWriters这个struct是包私有的 在package外不能创建变量
//entityAccessRegistry 是一个已经初始化好的包私有全局变量 在package外任然不能使用entityAccessRegistry
//在这里singleton的意思是，在这个package内可以直接使用这一个已经初始化好的全局变量，而不用重新创建一个变量
var entityAccessRegistry = &entityReaderWriters{
	protection: new(sync.RWMutex),
	accessors:  map[string]EntityReaderWriter{},
}

//MIME意为多目Internet邮件扩展，它设计的最初目的是为了在发送电子邮件时附加多媒体数据，让邮件客户程序能根据其类型进行处理。
//它被HTTP协议支持之后，它使得HTTP传输的不仅是普通的文本，而变得丰富多彩。
//在把输出结果传送到浏览器上的时候，浏览器必须启动适当的应用程序来处理这个输出文档
// entityReaderWriters associates MIME to an EntityReaderWriter
type entityReaderWriters struct {
	protection *sync.RWMutex
	accessors  map[string]EntityReaderWriter
}

func init() {
	RegisterEntityAccessor(MIME_JSON, entityJSONAccess{ContentType: MIME_JSON})
	RegisterEntityAccessor(MIME_XML, entityXMLAccess{ContentType: MIME_XML})
}

// RegisterEntityAccessor add/overrides the ReaderWriter for encoding content with this MIME type.
func RegisterEntityAccessor(mime string, erw EntityReaderWriter) {
	entityAccessRegistry.protection.Lock()
	defer entityAccessRegistry.protection.Unlock()
	entityAccessRegistry.accessors[mime] = erw
}

// AccessorAt returns the registered ReaderWriter for this MIME type.
func (r *entityReaderWriters) AccessorAt(mime string) (EntityReaderWriter, bool) {
	r.protection.RLock()
	defer r.protection.RUnlock()
	//accessors是一个map:  map[string]EntityReaderWriter ，key是各种MIME类型
	er, ok := r.accessors[mime] //comma，ok模式应用在map判定是否存在对应key的item
	if !ok {
		// retry with reverse lookup
		// more expensive but we are in an exceptional situation anyway
		for k, v := range r.accessors { // key,value := range map  ;  value,ispresentd:=range map[key]
			if strings.Contains(mime, k) { //判断mime中是否包含k
				return v, true
			}
		}
	}
	return er, ok
}

// entityXMLAccess is a EntityReaderWriter for XML encoding
type entityXMLAccess struct {
	// This is used for setting the Content-Type header when writing
	ContentType string
}

// Read unmarshalls the value from XML
func (e entityXMLAccess) Read(req *Request, v interface{}) error {
	return xml.NewDecoder(req.Request.Body).Decode(v)
}

// Write marshalls the value to JSON and set the Content-Type Header.
func (e entityXMLAccess) Write(resp *Response, status int, v interface{}) error {
	return writeXML(resp, status, e.ContentType, v)
}

// writeXML marshalls the value to JSON and set the Content-Type Header.
func writeXML(resp *Response, status int, contentType string, v interface{}) error {
	if v == nil {
		resp.WriteHeader(status)
		// do not write a nil representation
		return nil
	}
	if resp.prettyPrint {
		// pretty output must be created and written explicitly
		output, err := xml.MarshalIndent(v, " ", " ")
		if err != nil {
			return err
		}
		resp.Header().Set(HEADER_ContentType, contentType)
		resp.WriteHeader(status)
		_, err = resp.Write([]byte(xml.Header))
		if err != nil {
			return err
		}
		_, err = resp.Write(output)
		return err
	}
	// not-so-pretty
	resp.Header().Set(HEADER_ContentType, contentType)
	resp.WriteHeader(status)
	return xml.NewEncoder(resp).Encode(v)
}

// entityJSONAccess is a EntityReaderWriter for JSON encoding
type entityJSONAccess struct {
	// This is used for setting the Content-Type header when writing
	ContentType string
}

// Read unmarshalls the value from JSON
func (e entityJSONAccess) Read(req *Request, v interface{}) error {
	decoder := json.NewDecoder(req.Request.Body)
	decoder.UseNumber()
	return decoder.Decode(v)
}

// Write marshalls the value to JSON and set the Content-Type Header.
func (e entityJSONAccess) Write(resp *Response, status int, v interface{}) error {
	return writeJSON(resp, status, e.ContentType, v)
}

// write marshalls the value to JSON and set the Content-Type Header.
func writeJSON(resp *Response, status int, contentType string, v interface{}) error {
	if v == nil {
		resp.WriteHeader(status)
		// do not write a nil representation
		return nil
	}
	if resp.prettyPrint {
		// pretty output must be created and written explicitly
		output, err := json.MarshalIndent(v, " ", " ")
		if err != nil {
			return err
		}
		resp.Header().Set(HEADER_ContentType, contentType)
		resp.WriteHeader(status)
		_, err = resp.Write(output)
		return err
	}
	// not-so-pretty
	resp.Header().Set(HEADER_ContentType, contentType)
	resp.WriteHeader(status)
	return json.NewEncoder(resp).Encode(v)
}
