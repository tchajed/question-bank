// Package qti implements parsing of Canvas QTI zip export files.
//
// Canvas exports quizzes as IMS QTI 1.2 zip files. Each zip contains:
//   - imsmanifest.xml: index of resources in the package
//   - <id>/<id>.xml: the QTI assessment with questions
//   - <id>/assessment_meta.xml: Canvas-specific quiz metadata
package qti

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

// Manifest is the top-level IMS Content Package manifest (imsmanifest.xml).
type Manifest struct {
	XMLName   xml.Name           `xml:"manifest"`
	Resources []ManifestResource `xml:"resources>resource"`
}

// ManifestResource is one resource entry in the manifest.
type ManifestResource struct {
	Identifier string `xml:"identifier,attr"`
	Type       string `xml:"type,attr"`
	Href       string `xml:"href,attr"`
	Files      []struct {
		Href string `xml:"href,attr"`
	} `xml:"file"`
}

// AssessmentMeta holds Canvas-specific quiz metadata from assessment_meta.xml.
type AssessmentMeta struct {
	XMLName        xml.Name        `xml:"quiz"`
	Identifier     string          `xml:"identifier,attr"`
	Title          string          `xml:"title"`
	Description    string          `xml:"description"`
	ShuffleAnswers bool            `xml:"shuffle_answers"`
	QuizType       string          `xml:"quiz_type"`
	PointsPossible float64         `xml:"points_possible"`
	AllowedAttempts int            `xml:"allowed_attempts"`
	Assignment     MetaAssignment  `xml:"assignment"`
}

// MetaAssignment is the assignment sub-element of AssessmentMeta.
type MetaAssignment struct {
	Identifier     string  `xml:"identifier,attr"`
	Title          string  `xml:"title"`
	WorkflowState  string  `xml:"workflow_state"`
	PointsPossible float64 `xml:"points_possible"`
	GradingType    string  `xml:"grading_type"`
	SubmissionTypes string `xml:"submission_types"`
}

// Assessment is the root QTI assessment element in <id>.xml.
type Assessment struct {
	XMLName xml.Name       `xml:"questestinterop"`
	Root    AssessmentRoot `xml:"assessment"`
}

// AssessmentRoot is the <assessment> element inside <questestinterop>.
type AssessmentRoot struct {
	Ident    string          `xml:"ident,attr"`
	Title    string          `xml:"title,attr"`
	Metadata []MetadataField `xml:"qtimetadata>qtimetadatafield"`
	Items    []Item          `xml:"section>item"`
}

// MetadataField is a key/value pair in QTI metadata.
type MetadataField struct {
	Label string `xml:"fieldlabel"`
	Entry string `xml:"fieldentry"`
}

// Item represents a single question in the QTI assessment.
type Item struct {
	Ident    string          `xml:"ident,attr"`
	Title    string          `xml:"title,attr"`
	Metadata []MetadataField `xml:"itemmetadata>qtimetadata>qtimetadatafield"`

	// Presentation holds the question text and response choices.
	Presentation Presentation `xml:"presentation"`

	// ResProcessing defines the scoring rules.
	ResProcessing ResProcessing `xml:"resprocessing"`

	// ItemFeedback holds feedback text keyed by ident.
	ItemFeedback []ItemFeedback `xml:"itemfeedback"`
}

// QuestionType returns the Canvas question type from item metadata
// (e.g. "multiple_choice_question", "true_false_question", "multiple_answers_question").
func (it *Item) QuestionType() string {
	for _, f := range it.Metadata {
		if f.Label == "question_type" {
			return f.Entry
		}
	}
	return ""
}

// PointsPossible returns the point value from item metadata.
func (it *Item) PointsPossible() string {
	for _, f := range it.Metadata {
		if f.Label == "points_possible" {
			return f.Entry
		}
	}
	return ""
}

// Presentation contains the question stem and response options.
type Presentation struct {
	// Material holds the question text (HTML or plain text).
	Material Material `xml:"material"`
	// ResponseLid is the response container for choice questions.
	ResponseLid *ResponseLid `xml:"response_lid"`
}

// Material holds text content.
type Material struct {
	MatText MatText `xml:"mattext"`
}

// MatText is a text element with an optional MIME type.
type MatText struct {
	TextType string `xml:"texttype,attr"`
	Text     string `xml:",chardata"`
}

// ResponseLid is a response container for single- or multiple-select questions.
type ResponseLid struct {
	Ident       string          `xml:"ident,attr"`
	RCardinality string         `xml:"rcardinality,attr"` // "Single" or "Multiple"
	Choices     []ResponseLabel `xml:"render_choice>response_label"`
}

// ResponseLabel is one answer choice.
type ResponseLabel struct {
	Ident    string   `xml:"ident,attr"`
	Material Material `xml:"material"`
}

// ResProcessing holds the scoring/response processing rules for an item.
type ResProcessing struct {
	Outcomes       []DecVar        `xml:"outcomes>decvar"`
	RespConditions []RespCondition `xml:"respcondition"`
}

// DecVar is a declared score variable.
type DecVar struct {
	VarName  string `xml:"varname,attr"`
	VarType  string `xml:"vartype,attr"`
	MinValue string `xml:"minvalue,attr"`
	MaxValue string `xml:"maxvalue,attr"`
}

// RespCondition is a conditional scoring rule.
type RespCondition struct {
	Continue     string       `xml:"continue,attr"`
	ConditionVar ConditionVar `xml:"conditionvar"`
	SetVar       *SetVar      `xml:"setvar"`
	DisplayFeedback []DisplayFeedback `xml:"displayfeedback"`
}

// ConditionVar is the condition for a respcondition.
type ConditionVar struct {
	// Other is true when the <other/> element is present (matches any response).
	Other bool `xml:"-"`
	// VarEquals are direct equality conditions.
	VarEquals []VarEqual `xml:"varequal"`
	// And groups conditions with logical AND.
	And *AndCondition `xml:"and"`
}

// UnmarshalXML implements xml.Unmarshaler to detect the <other/> element.
func (c *ConditionVar) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	// Use a local type to avoid recursion.
	type conditionVarAlias struct {
		VarEquals []VarEqual   `xml:"varequal"`
		And       *AndCondition `xml:"and"`
		Other     *struct{}    `xml:"other"`
	}
	var alias conditionVarAlias
	if err := d.DecodeElement(&alias, &start); err != nil {
		return err
	}
	c.VarEquals = alias.VarEquals
	c.And = alias.And
	c.Other = alias.Other != nil
	return nil
}

// AndCondition groups conditions with logical AND (used in multiple_answers_question).
type AndCondition struct {
	VarEquals []VarEqual      `xml:"varequal"`
	Nots      []NotCondition  `xml:"not"`
}

// NotCondition wraps a negated condition.
type NotCondition struct {
	VarEquals []VarEqual `xml:"varequal"`
}

// VarEqual is an equality check against a response variable.
type VarEqual struct {
	RespIdent string `xml:"respident,attr"`
	Value     string `xml:",chardata"`
}

// SetVar sets a score variable.
type SetVar struct {
	VarName string `xml:"varname,attr"`
	Action  string `xml:"action,attr"`
	Value   string `xml:",chardata"`
}

// DisplayFeedback references feedback to show.
type DisplayFeedback struct {
	FeedbackType string `xml:"feedbacktype,attr"`
	LinkRefId    string `xml:"linkrefid,attr"`
}

// ItemFeedback holds feedback text for a response outcome.
type ItemFeedback struct {
	Ident    string   `xml:"ident,attr"`
	Material Material `xml:"flow_mat>material"`
}

// Quiz is the fully parsed contents of a Canvas QTI zip export.
type Quiz struct {
	Meta       AssessmentMeta
	Assessment Assessment
}

// ParseZip reads a Canvas QTI zip file and returns the parsed Quiz.
func ParseZip(path string) (*Quiz, error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return nil, fmt.Errorf("open zip: %w", err)
	}
	defer r.Close()

	var manifestData, metaData, assessmentData []byte

	for _, f := range r.File {
		name := f.Name
		data, err := readZipFile(f)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", name, err)
		}
		switch {
		case name == "imsmanifest.xml":
			manifestData = data
		case strings.HasSuffix(name, "/assessment_meta.xml"):
			metaData = data
		case strings.HasSuffix(name, ".xml") && !strings.HasSuffix(name, "/assessment_meta.xml") && name != "imsmanifest.xml":
			assessmentData = data
		}
	}

	if manifestData == nil {
		return nil, fmt.Errorf("imsmanifest.xml not found in zip")
	}
	if metaData == nil {
		return nil, fmt.Errorf("assessment_meta.xml not found in zip")
	}
	if assessmentData == nil {
		return nil, fmt.Errorf("assessment XML not found in zip")
	}

	var quiz Quiz
	if err := xml.Unmarshal(metaData, &quiz.Meta); err != nil {
		return nil, fmt.Errorf("parse assessment_meta.xml: %w", err)
	}
	if err := xml.Unmarshal(assessmentData, &quiz.Assessment); err != nil {
		return nil, fmt.Errorf("parse assessment XML: %w", err)
	}
	return &quiz, nil
}

func readZipFile(f *zip.File) ([]byte, error) {
	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	return io.ReadAll(rc)
}
