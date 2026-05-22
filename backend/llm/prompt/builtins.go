package prompt

// Summarize generates a concise summary of the provided text.
var Summarize = &Template{
	Name:   "summarize",
	System: "You are a precise summarization assistant. Produce clear, concise summaries.",
	User:   "Summarize the following text concisely:\n\n{{.Text}}",
}

// Translate converts text between languages.
var Translate = &Template{
	Name:   "translate",
	System: "You are a professional translator. Translate accurately while preserving tone and meaning.",
	User:   "Translate the following text from {{.SourceLang}} to {{.TargetLang}}:\n\n{{.Text}}",
}

// Analyze provides analytical insights on the given data.
var Analyze = &Template{
	Name:   "analyze",
	System: "You are a data analysis expert. Provide structured, actionable insights.",
	User:   "Analyze the following data and provide insights:\n\n{{.Data}}",
}

// QA answers questions based on the provided context.
var QA = &Template{
	Name:   "qa",
	System: "Answer questions accurately based only on the provided context. If the context does not contain enough information, say so.",
	User:   "Context:\n{{.Context}}\n\nQuestion: {{.Question}}",
}

// Sentiment analyzes the emotional tone of text.
var Sentiment = &Template{
	Name:   "sentiment",
	System: "You are a sentiment analysis expert. Classify sentiment and explain your reasoning.",
	User:   "Analyze the sentiment of the following text:\n\n{{.Text}}",
}

// Classify categorizes text into the specified categories.
var Classify = &Template{
	Name:   "classify",
	System: "You are a text classification expert. Classify text accurately into the given categories.",
	User:   "Classify the following text into categories ({{.Categories}}):\n\n{{.Text}}",
}

// Extract pulls key information from text in a structured format.
var Extract = &Template{
	Name:   "extract",
	System: "You are an information extraction expert. Extract key information in a structured format.",
	User:   "Extract key information from the following text:\n\n{{.Text}}",
}
