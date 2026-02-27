package qti

import (
	"archive/zip"
	"crypto/rand"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// NewChoice is one answer option for a new quiz item.
type NewChoice struct {
	// ID is the choice identifier. If empty, a unique ID is assigned.
	ID      string
	Text    string
	Correct bool
}

// ItemType is the Canvas question type.
type ItemType string

const (
	TrueFalseQuestion     ItemType = "true_false_question"
	MultipleChoiceQuestion  ItemType = "multiple_choice_question"
	MultipleAnswersQuestion ItemType = "multiple_answers_question"
)

// NewItem describes a single quiz question to create.
type NewItem struct {
	// ID is the item identifier. If empty, a unique ID is assigned.
	ID    string
	Title string
	// Text is the question stem in HTML.
	Text    string
	Type    ItemType
	Points  float64
	Choices []NewChoice
	// Feedback text in HTML. GeneralFeedback is always displayed.
	// CorrectFeedback is shown on a correct response.
	// IncorrectFeedback is shown on an incorrect response.
	GeneralFeedback   string
	CorrectFeedback   string
	IncorrectFeedback string
}

// NewQuiz describes a quiz to create as a Canvas QTI zip file.
type NewQuiz struct {
	// ID is the quiz identifier. If empty, a unique ID is assigned.
	ID             string
	Title          string
	Description    string
	PointsPossible float64
	AllowedAttempts int
	// QuizType is the Canvas quiz type (default: "assignment").
	QuizType string
	Items    []NewItem
}

// WriteZip creates a Canvas QTI zip file at path from quiz.
func WriteZip(path string, quiz *NewQuiz) (err error) {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := f.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	w := zip.NewWriter(f)
	defer func() {
		if cerr := w.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	quizID := quiz.ID
	if quizID == "" {
		quizID = generateID()
	}
	metaID := generateID()
	dir := quizID + "/"

	manifestXML, err := marshalXML(buildManifest(quizID, metaID, dir))
	if err != nil {
		return fmt.Errorf("build manifest: %w", err)
	}
	if err := writeZipEntry(w, "imsmanifest.xml", manifestXML); err != nil {
		return fmt.Errorf("write manifest: %w", err)
	}

	metaXML, err := marshalXML(buildMeta(quizID, quiz))
	if err != nil {
		return fmt.Errorf("build assessment_meta: %w", err)
	}
	if err := writeZipEntry(w, dir+"assessment_meta.xml", metaXML); err != nil {
		return fmt.Errorf("write assessment_meta: %w", err)
	}

	assessXML, err := marshalXML(buildAssessment(quizID, quiz))
	if err != nil {
		return fmt.Errorf("build assessment XML: %w", err)
	}
	if err := writeZipEntry(w, dir+quizID+".xml", assessXML); err != nil {
		return fmt.Errorf("write assessment XML: %w", err)
	}

	return nil
}

// ---- manifest ----

type wManifest struct {
	XMLName    xml.Name           `xml:"manifest"`
	Identifier string             `xml:"identifier,attr"`
	Resources  wManifestResources `xml:"resources"`
}

type wManifestResources struct {
	Resources []wManifestResource `xml:"resource"`
}

type wManifestResource struct {
	Identifier string       `xml:"identifier,attr"`
	Type       string       `xml:"type,attr"`
	Href       string       `xml:"href,attr,omitempty"`
	Files      []wResFile   `xml:"file"`
	Dependency *wDependency `xml:"dependency"`
}

type wResFile struct {
	Href string `xml:"href,attr"`
}

type wDependency struct {
	IdentifierRef string `xml:"identifierref,attr"`
}

func buildManifest(quizID, metaID, dir string) wManifest {
	return wManifest{
		Identifier: generateID(),
		Resources: wManifestResources{
			Resources: []wManifestResource{
				{
					Identifier: quizID,
					Type:       "imsqti_xmlv1p2",
					Files:      []wResFile{{Href: dir + quizID + ".xml"}},
					Dependency: &wDependency{IdentifierRef: metaID},
				},
				{
					Identifier: metaID,
					Type:       "associatedcontent/imscc_xmlv1p1/learning-application-resource",
					Href:       dir + "assessment_meta.xml",
					Files:      []wResFile{{Href: dir + "assessment_meta.xml"}},
				},
			},
		},
	}
}

// ---- assessment_meta.xml ----

type wQuizMeta struct {
	XMLName         xml.Name `xml:"quiz"`
	Identifier      string   `xml:"identifier,attr"`
	Title           string   `xml:"title"`
	Description     string   `xml:"description"`
	ShuffleAnswers  bool     `xml:"shuffle_answers"`
	QuizType        string   `xml:"quiz_type"`
	PointsPossible  float64  `xml:"points_possible"`
	AllowedAttempts int      `xml:"allowed_attempts"`
}

func buildMeta(quizID string, quiz *NewQuiz) wQuizMeta {
	quizType := quiz.QuizType
	if quizType == "" {
		quizType = "assignment"
	}
	allowed := quiz.AllowedAttempts
	if allowed == 0 {
		allowed = 1
	}
	return wQuizMeta{
		Identifier:      quizID,
		Title:           quiz.Title,
		Description:     quiz.Description,
		QuizType:        quizType,
		PointsPossible:  quiz.PointsPossible,
		AllowedAttempts: allowed,
	}
}

// ---- assessment XML ----

type wQTI struct {
	XMLName    xml.Name    `xml:"questestinterop"`
	Assessment wAssessment `xml:"assessment"`
}

type wAssessment struct {
	Ident    string   `xml:"ident,attr"`
	Title    string   `xml:"title,attr"`
	Metadata wQtiMeta `xml:"qtimetadata"`
	Section  wSection `xml:"section"`
}

type wQtiMeta struct {
	Fields []MetadataField `xml:"qtimetadatafield"`
}

type wSection struct {
	Ident string  `xml:"ident,attr"`
	Items []wItem `xml:"item"`
}

type wItem struct {
	Ident        string         `xml:"ident,attr"`
	Title        string         `xml:"title,attr"`
	Metadata     wItemMeta      `xml:"itemmetadata"`
	Presentation wPresentation  `xml:"presentation"`
	ResProc      wResProc       `xml:"resprocessing"`
	Feedback     []wItemFB      `xml:"itemfeedback"`
}

type wItemMeta struct {
	QtiMeta wQtiMeta `xml:"qtimetadata"`
}

type wPresentation struct {
	Material    wMaterial    `xml:"material"`
	ResponseLid wResponseLid `xml:"response_lid"`
}

type wMaterial struct {
	MatText wMatText `xml:"mattext"`
}

type wMatText struct {
	TextType string `xml:"texttype,attr"`
	Text     string `xml:",chardata"`
}

type wResponseLid struct {
	Ident        string        `xml:"ident,attr"`
	RCardinality string        `xml:"rcardinality,attr"`
	RenderChoice wRenderChoice `xml:"render_choice"`
}

type wRenderChoice struct {
	Labels []wResponseLabel `xml:"response_label"`
}

type wResponseLabel struct {
	Ident    string    `xml:"ident,attr"`
	Material wMaterial `xml:"material"`
}

type wResProc struct {
	Outcomes []wDecVar   `xml:"outcomes>decvar"`
	Conds    []wRespCond `xml:"respcondition"`
}

type wDecVar struct {
	VarName  string `xml:"varname,attr"`
	VarType  string `xml:"vartype,attr"`
	MinValue string `xml:"minvalue,attr"`
	MaxValue string `xml:"maxvalue,attr"`
}

type wRespCond struct {
	Continue     string       `xml:"continue,attr"`
	ConditionVar wCondVar     `xml:"conditionvar"`
	SetVar       *wSetVar     `xml:"setvar"`
	DisplayFB    []wDisplayFB `xml:"displayfeedback"`
}

// wCondVar handles the <other/> empty element via MarshalXML.
type wCondVar struct {
	isOther   bool
	varEquals []wVarEqual
	and       *wAnd
}

func (c wCondVar) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if err := e.EncodeToken(start); err != nil {
		return err
	}
	if c.isOther {
		other := xml.StartElement{Name: xml.Name{Local: "other"}}
		if err := e.EncodeToken(other); err != nil {
			return err
		}
		if err := e.EncodeToken(other.End()); err != nil {
			return err
		}
	}
	for _, ve := range c.varEquals {
		if err := e.EncodeElement(ve, xml.StartElement{Name: xml.Name{Local: "varequal"}}); err != nil {
			return err
		}
	}
	if c.and != nil {
		if err := e.EncodeElement(c.and, xml.StartElement{Name: xml.Name{Local: "and"}}); err != nil {
			return err
		}
	}
	return e.EncodeToken(start.End())
}

type wVarEqual struct {
	RespIdent string `xml:"respident,attr"`
	Value     string `xml:",chardata"`
}

type wAnd struct {
	VarEquals []wVarEqual `xml:"varequal"`
	Nots      []wNot      `xml:"not"`
}

type wNot struct {
	VarEquals []wVarEqual `xml:"varequal"`
}

type wSetVar struct {
	VarName string `xml:"varname,attr"`
	Action  string `xml:"action,attr"`
	Value   string `xml:",chardata"`
}

type wDisplayFB struct {
	FeedbackType string `xml:"feedbacktype,attr"`
	LinkRefId    string `xml:"linkrefid,attr"`
}

type wItemFB struct {
	Ident   string    `xml:"ident,attr"`
	FlowMat wFlowMat  `xml:"flow_mat"`
}

type wFlowMat struct {
	Material wMaterial `xml:"material"`
}

func buildAssessment(quizID string, quiz *NewQuiz) wQTI {
	allowed := quiz.AllowedAttempts
	if allowed == 0 {
		allowed = 1
	}
	items := make([]wItem, len(quiz.Items))
	for i := range quiz.Items {
		items[i] = buildItem(&quiz.Items[i])
	}
	return wQTI{
		Assessment: wAssessment{
			Ident: quizID,
			Title: quiz.Title,
			Metadata: wQtiMeta{Fields: []MetadataField{
				{Label: "cc_maxattempts", Entry: strconv.Itoa(allowed)},
			}},
			Section: wSection{Ident: "root_section", Items: items},
		},
	}
}

func buildItem(item *NewItem) wItem {
	id := item.ID
	if id == "" {
		id = generateID()
	}
	qtype := string(item.Type)
	rcard := "Single"
	if item.Type == MultipleAnswersQuestion {
		rcard = "Multiple"
	}
	points := item.Points
	if points == 0 {
		points = 1.0
	}

	// Assign choice IDs where missing.
	choices := make([]NewChoice, len(item.Choices))
	copy(choices, item.Choices)
	for i := range choices {
		if choices[i].ID == "" {
			choices[i].ID = generateID()
		}
	}

	labels := make([]wResponseLabel, len(choices))
	for i, c := range choices {
		labels[i] = wResponseLabel{
			Ident:    c.ID,
			Material: wMaterial{MatText: wMatText{TextType: "text/plain", Text: c.Text}},
		}
	}

	origIDs := make([]string, len(choices))
	for i, c := range choices {
		origIDs[i] = c.ID
	}
	fields := []MetadataField{
		{Label: "question_type", Entry: qtype},
		{Label: "points_possible", Entry: fmt.Sprintf("%.1f", points)},
		{Label: "original_answer_ids", Entry: strings.Join(origIDs, ",")},
		{Label: "assessment_question_identifierref", Entry: generateID()},
	}

	var feedbacks []wItemFB
	if item.GeneralFeedback != "" {
		feedbacks = append(feedbacks, wFeedback("general_fb", item.GeneralFeedback))
	}
	if item.CorrectFeedback != "" {
		feedbacks = append(feedbacks, wFeedback("correct_fb", item.CorrectFeedback))
	}
	if item.IncorrectFeedback != "" {
		feedbacks = append(feedbacks, wFeedback("general_incorrect_fb", item.IncorrectFeedback))
	}

	return wItem{
		Ident:    id,
		Title:    item.Title,
		Metadata: wItemMeta{QtiMeta: wQtiMeta{Fields: fields}},
		Presentation: wPresentation{
			Material: wMaterial{MatText: wMatText{TextType: "text/html", Text: item.Text}},
			ResponseLid: wResponseLid{
				Ident:        "response1",
				RCardinality: rcard,
				RenderChoice: wRenderChoice{Labels: labels},
			},
		},
		ResProc:  buildResProc(item, choices),
		Feedback: feedbacks,
	}
}

func buildResProc(item *NewItem, choices []NewChoice) wResProc {
	outcomes := []wDecVar{
		{VarName: "SCORE", VarType: "Decimal", MinValue: "0", MaxValue: "100"},
	}
	var conds []wRespCond

	switch item.Type {
	case TrueFalseQuestion, MultipleChoiceQuestion:
		var correctID string
		for _, c := range choices {
			if c.Correct {
				correctID = c.ID
				break
			}
		}
		if item.GeneralFeedback != "" {
			conds = append(conds, wRespCond{
				Continue:     "Yes",
				ConditionVar: wCondVar{isOther: true},
				DisplayFB:    []wDisplayFB{{FeedbackType: "Response", LinkRefId: "general_fb"}},
			})
		}
		correct := wRespCond{
			Continue:     "No",
			ConditionVar: wCondVar{varEquals: []wVarEqual{{RespIdent: "response1", Value: correctID}}},
			SetVar:       &wSetVar{VarName: "SCORE", Action: "Set", Value: "100"},
		}
		if item.CorrectFeedback != "" {
			correct.DisplayFB = []wDisplayFB{{FeedbackType: "Response", LinkRefId: "correct_fb"}}
		}
		conds = append(conds, correct)
		if item.IncorrectFeedback != "" {
			conds = append(conds, wRespCond{
				Continue:     "Yes",
				ConditionVar: wCondVar{isOther: true},
				DisplayFB:    []wDisplayFB{{FeedbackType: "Response", LinkRefId: "general_incorrect_fb"}},
			})
		}

	case MultipleAnswersQuestion:
		var varEquals []wVarEqual
		var nots []wNot
		for _, c := range choices {
			if c.Correct {
				varEquals = append(varEquals, wVarEqual{RespIdent: "response1", Value: c.ID})
			} else {
				nots = append(nots, wNot{VarEquals: []wVarEqual{{RespIdent: "response1", Value: c.ID}}})
			}
		}
		conds = append(conds, wRespCond{
			Continue:     "No",
			ConditionVar: wCondVar{and: &wAnd{VarEquals: varEquals, Nots: nots}},
			SetVar:       &wSetVar{VarName: "SCORE", Action: "Set", Value: "100"},
		})
		if item.IncorrectFeedback != "" {
			conds = append(conds, wRespCond{
				Continue:     "Yes",
				ConditionVar: wCondVar{isOther: true},
				DisplayFB:    []wDisplayFB{{FeedbackType: "Response", LinkRefId: "general_incorrect_fb"}},
			})
		}
	}

	return wResProc{Outcomes: outcomes, Conds: conds}
}

func wFeedback(ident, html string) wItemFB {
	return wItemFB{
		Ident:   ident,
		FlowMat: wFlowMat{Material: wMaterial{MatText: wMatText{TextType: "text/html", Text: html}}},
	}
}

// ---- helpers ----

func generateID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic(fmt.Sprintf("qti: failed to generate ID: %v", err))
	}
	return "g" + hex.EncodeToString(b)
}

func marshalXML(v any) ([]byte, error) {
	data, err := xml.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, err
	}
	return append([]byte(xml.Header), data...), nil
}

func writeZipEntry(w *zip.Writer, name string, data []byte) error {
	f, err := w.Create(name)
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	return err
}
