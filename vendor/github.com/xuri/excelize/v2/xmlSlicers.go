// Copyright 2016 - 2024 The excelize Authors. All rights reserved. Use of
// this source code is governed by a BSD-style license that can be found in
// the LICENSE file.
//
// Package excelize providing a set of functions that allow you to write to and
// read from XLAM / XLSM / XLSX / XLTM / XLTX files. Supports reading and
// writing spreadsheet documents generated by Microsoft Excel™ 2007 and later.
// Supports complex components by high compatibility, and provided streaming
// API for generating or reading data from a worksheet with huge amounts of
// data. This library needs Go version 1.18 or later.

package excelize

import "encoding/xml"

// xlsxSlicers directly maps the slicers element that specifies a slicer view on
// the worksheet.
type xlsxSlicers struct {
	XMLName   xml.Name     `xml:"http://schemas.microsoft.com/office/spreadsheetml/2009/9/main slicers"`
	XMLNSXMC  string       `xml:"xmlns:mc,attr"`
	XMLNSX    string       `xml:"xmlns:x,attr"`
	XMLNSXR10 string       `xml:"xmlns:xr10,attr"`
	Slicer    []xlsxSlicer `xml:"slicer"`
}

// xlsxSlicer is a complex type that specifies a slicer view.
type xlsxSlicer struct {
	Name           string `xml:"name,attr"`
	XR10UID        string `xml:"xr10:uid,attr,omitempty"`
	Cache          string `xml:"cache,attr"`
	Caption        string `xml:"caption,attr,omitempty"`
	StartItem      *int   `xml:"startItem,attr"`
	ColumnCount    *int   `xml:"columnCount,attr"`
	ShowCaption    *bool  `xml:"showCaption,attr"`
	Level          int    `xml:"level,attr,omitempty"`
	Style          string `xml:"style,attr,omitempty"`
	LockedPosition bool   `xml:"lockedPosition,attr,omitempty"`
	RowHeight      int    `xml:"rowHeight,attr"`
}

// slicerCacheDefinition directly maps the slicerCacheDefinition element that
// specifies a slicer cache.
type xlsxSlicerCacheDefinition struct {
	XMLName     xml.Name                    `xml:"http://schemas.microsoft.com/office/spreadsheetml/2009/9/main slicerCacheDefinition"`
	XMLNSXMC    string                      `xml:"xmlns:mc,attr"`
	XMLNSX      string                      `xml:"xmlns:x,attr"`
	XMLNSX15    string                      `xml:"xmlns:x15,attr,omitempty"`
	XMLNSXR10   string                      `xml:"xmlns:xr10,attr"`
	Name        string                      `xml:"name,attr"`
	XR10UID     string                      `xml:"xr10:uid,attr,omitempty"`
	SourceName  string                      `xml:"sourceName,attr"`
	PivotTables *xlsxSlicerCachePivotTables `xml:"pivotTables"`
	Data        *xlsxSlicerCacheData        `xml:"data"`
	ExtLst      *xlsxExtLst                 `xml:"extLst"`
}

// xlsxSlicerCachePivotTables is a complex type that specifies a group of
// pivotTable elements that specify the PivotTable views that are filtered by
// the slicer cache.
type xlsxSlicerCachePivotTables struct {
	PivotTable []xlsxSlicerCachePivotTable `xml:"pivotTable"`
}

// xlsxSlicerCachePivotTable is a complex type that specifies a PivotTable view
// filtered by a slicer cache.
type xlsxSlicerCachePivotTable struct {
	TabID int    `xml:"tabId,attr"`
	Name  string `xml:"name,attr"`
}

// xlsxSlicerCacheData is a complex type that specifies a data source for the
// slicer cache.
type xlsxSlicerCacheData struct {
	OLAP    *xlsxInnerXML           `xml:"olap"`
	Tabular *xlsxTabularSlicerCache `xml:"tabular"`
}

// xlsxTabularSlicerCache is a complex type that specifies non-OLAP slicer items
// that are cached within this slicer cache and properties of the slicer cache
// specific to non-OLAP slicer items.
type xlsxTabularSlicerCache struct {
	PivotCacheID   int                          `xml:"pivotCacheId,attr"`
	SortOrder      string                       `xml:"sortOrder,attr,omitempty"`
	CustomListSort *bool                        `xml:"customListSort,attr"`
	ShowMissing    *bool                        `xml:"showMissing,attr"`
	CrossFilter    string                       `xml:"crossFilter,attr,omitempty"`
	Items          *xlsxTabularSlicerCacheItems `xml:"items"`
	ExtLst         *xlsxExtLst                  `xml:"extLst"`
}

// xlsxTabularSlicerCacheItems is a complex type that specifies non-OLAP slicer
// items that are cached within this slicer cache.
type xlsxTabularSlicerCacheItems struct {
	Count int                          `xml:"count,attr,omitempty"`
	I     []xlsxTabularSlicerCacheItem `xml:"i"`
}

// xlsxTabularSlicerCacheItem is a complex type that specifies a non-OLAP slicer
// item that is cached within this slicer cache.
type xlsxTabularSlicerCacheItem struct {
	X  int  `xml:"x,attr"`
	S  bool `xml:"s,attr,omitempty"`
	ND bool `xml:"nd,attr,omitempty"`
}

// xlsxTableSlicerCache specifies a table data source for the slicer cache.
type xlsxTableSlicerCache struct {
	XMLName        xml.Name    `xml:"x15:tableSlicerCache"`
	TableID        int         `xml:"tableId,attr"`
	Column         int         `xml:"column,attr"`
	SortOrder      string      `xml:"sortOrder,attr,omitempty"`
	CustomListSort *bool       `xml:"customListSort,attr"`
	CrossFilter    string      `xml:"crossFilter,attr,omitempty"`
	ExtLst         *xlsxExtLst `xml:"extLst"`
}

// xlsxX14SlicerList specifies a list of slicer.
type xlsxX14SlicerList struct {
	XMLName xml.Name         `xml:"x14:slicerList"`
	Slicer  []*xlsxX14Slicer `xml:"x14:slicer"`
}

// xlsxX14Slicer specifies a slicer view,
type xlsxX14Slicer struct {
	XMLName xml.Name `xml:"x14:slicer"`
	RID     string   `xml:"r:id,attr"`
}

// xlsxX14SlicerCaches directly maps the x14:slicerCache element.
type xlsxX14SlicerCaches struct {
	XMLName xml.Name `xml:"x14:slicerCaches"`
	XMLNS   string   `xml:"xmlns:x14,attr"`
	Content string   `xml:",innerxml"`
}

// xlsxX15SlicerCaches directly maps the x14:slicerCache element.
type xlsxX14SlicerCache struct {
	XMLName xml.Name `xml:"x14:slicerCache"`
	RID     string   `xml:"r:id,attr"`
}

// xlsxX15SlicerCaches directly maps the x15:slicerCaches element.
type xlsxX15SlicerCaches struct {
	XMLName xml.Name `xml:"x15:slicerCaches"`
	XMLNS   string   `xml:"xmlns:x14,attr"`
	Content string   `xml:",innerxml"`
}

// decodeTableSlicerCache defines the structure used to parse the
// x15:tableSlicerCache element of the table slicer cache.
type decodeTableSlicerCache struct {
	XMLName   xml.Name `xml:"tableSlicerCache"`
	TableID   int      `xml:"tableId,attr"`
	Column    int      `xml:"column,attr"`
	SortOrder string   `xml:"sortOrder,attr"`
}

// decodeSlicerList defines the structure used to parse the x14:slicerList
// element of a list of slicer.
type decodeSlicerList struct {
	XMLName xml.Name        `xml:"slicerList"`
	Slicer  []*decodeSlicer `xml:"slicer"`
}

// decodeSlicer defines the structure used to parse the x14:slicer element of a
// slicer.
type decodeSlicer struct {
	RID string `xml:"id,attr"`
}

// decodeSlicerCaches defines the structure used to parse the
// x14:slicerCaches and x15:slicerCaches element of a slicer cache.
type decodeSlicerCaches struct {
	XMLName xml.Name `xml:"slicerCaches"`
	Content string   `xml:",innerxml"`
}

// xlsxTimelines is a mechanism for filtering data in pivot table views, cube
// functions and charts based on non-worksheet pivot tables. In the case of
// using OLAP Timeline source data, a Timeline is based on a key attribute of
// an OLAP hierarchy. In the case of using native Timeline source data, a
// Timeline is based on a data table column.
type xlsxTimelines struct {
	XMLName   xml.Name       `xml:"http://schemas.microsoft.com/office/spreadsheetml/2010/11/main timelines"`
	XMLNSXMC  string         `xml:"xmlns:mc,attr"`
	XMLNSX    string         `xml:"xmlns:x,attr"`
	XMLNSXR10 string         `xml:"xmlns:xr10,attr"`
	Timeline  []xlsxTimeline `xml:"timeline"`
}

// xlsxTimeline is timeline view specifies the display of a timeline on a
// worksheet.
type xlsxTimeline struct {
	Name                    string `xml:"name,attr"`
	XR10UID                 string `xml:"xr10:uid,attr,omitempty"`
	Cache                   string `xml:"cache,attr"`
	Caption                 string `xml:"caption,attr,omitempty"`
	ShowHeader              *bool  `xml:"showHeader,attr"`
	ShowSelectionLabel      *bool  `xml:"showSelectionLabel,attr"`
	ShowTimeLevel           *bool  `xml:"showTimeLevel,attr"`
	ShowHorizontalScrollbar *bool  `xml:"showHorizontalScrollbar,attr"`
	Level                   int    `xml:"level,attr"`
	SelectionLevel          int    `xml:"selectionLevel,attr"`
	ScrollPosition          string `xml:"scrollPosition,attr,omitempty"`
	Style                   string `xml:"style,attr,omitempty"`
}
