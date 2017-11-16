package main

import (
	"fmt"
	"reflect"
	"github.com/6thplaneta/hermes"
	"time"
)

type File struct {
	Id            int       `json:"id" hermes:"dbspace:files"`
	File_Path     string    `json:"file_path" validate:"required" hermes:"editable,searchable,index"`
	Creation_Date time.Time `json:"creation_date" hermes:"type:time"`
	Is_Deleted    bool      `json:"is_deleted" hermes:"editable,index"`
}
type FileCollection struct {
	*hermes.Collection
}

func NewFileCollection() (*FileCollection, error) {
	coll, err := hermes.NewDBCollection(&File{}, application.DataSrc)

	typ := reflect.TypeOf(File{})
	OColl := &FileCollection{coll}
	hermes.CollectionsMap[typ] = OColl

	return OColl, err
}

// getCollectionn
func (col *FileCollection) List(token string, params *hermes.Params, pg *hermes.Paging, populate, project string) (interface{}, error) {
	paramsList := params.List
	paramsList["Is_Deleted"] = hermes.Filter{Type: "exact", Value: false, FieldType: "bool"}

	obj, err := col.Collection.List(token, params, pg, populate, project)
	if err != nil {
		return nil, err
	}

	return obj, nil
}

func (col *FileCollection) VirtualDelete(token string, id int) error {

	_, err := col.DataSrc.DB.Exec(fmt.Sprintf("update %s  set is_deleted=true where Id= %d", col.Dbspace, id))
	if err != nil {
		return err
	}

	return nil
}
