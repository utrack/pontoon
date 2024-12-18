package docmerge

import (
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

// Stage represents a single documentation processing stage
type Stage interface {
	Process(doc *v3.Document) error
}

// Pipeline processes documentation through multiple stages
type Pipeline struct {
	stages []Stage
}

// NewPipeline creates a new documentation processing pipeline
func NewPipeline(stages ...Stage) *Pipeline {
	return &Pipeline{
		stages: stages,
	}
}

// Process runs the document through all stages
func (p *Pipeline) Process(doc *v3.Document) error {
	for _, stage := range p.stages {
		if err := stage.Process(doc); err != nil {
			return err
		}
	}
	return nil
}

// CommentProcessor processes Go-style documentation comments
type CommentProcessor struct {
	// TODO: add configuration options
}

// NewCommentProcessor creates a new comment processor
func NewCommentProcessor() *CommentProcessor {
	return &CommentProcessor{}
}

// Process implements Stage interface
func (p *CommentProcessor) Process(doc *v3.Document) error {
	// TODO: implement comment processing
	// 1. Extract deprecation status from comments
	// 2. Mark corresponding OpenAPI elements as deprecated
	return nil
}
