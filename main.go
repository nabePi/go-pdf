package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/pdfcpu/pdfcpu/pkg/api"
)

type TextField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Locked bool   `json:"locked"`
}

type Form struct {
	TextFields []TextField `json:"textfield"`
}

type PDFForm struct {
	Forms []Form `json:"forms"`
}

func main() {
	http.HandleFunc("/replay-sjk", replaySJKHandler)
	fmt.Println("Server is running on port 8080")
	http.ListenAndServe(":8080", nil)
}

func replaySJKHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Unable to read request body", http.StatusBadRequest)
		return
	}
	var formDataMap map[string]string
	err = json.Unmarshal(body, &formDataMap)
	if err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	pdfForm := mapFormDataToPDFForm(formDataMap)

	jsonData, err := json.Marshal(pdfForm)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return
	}

	dataReader := bytes.NewReader(jsonData)

	// Fill the form
	err = fillPDFForm("template_replay_form_sjk.pdf", "filled_form.pdf", dataReader)
	if err != nil {
		fmt.Println("Error filling form:", err)
		http.Error(w, "Failed to fill form", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Form filled successfully"}`))
}

func fillPDFForm(templatePath, outputPath string, data io.Reader) error {
	// Open the PDF template
	templateFile, err := os.Open(templatePath)
	if err != nil {
		return fmt.Errorf("failed to open PDF file: %w", err)
	}
	defer templateFile.Close()

	// Create the output PDF file
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output PDF file: %w", err)
	}
	defer outFile.Close()

	// Fill the form fields
	err = api.FillForm(templateFile, data, outFile, nil)
	if err != nil {
		return fmt.Errorf("failed to fill form: %w", err)
	}

	return nil
}

func mapFormDataToPDFForm(formData map[string]string) PDFForm {
	var textFields []TextField

	for key, value := range formData {
		textField := TextField{
			Name:   key,
			Value:  value,
			Locked: true,
		}
		textFields = append(textFields, textField)
	}

	form := Form{
		TextFields: textFields,
	}

	pdfForm := PDFForm{
		Forms: []Form{form},
	}

	return pdfForm
}
