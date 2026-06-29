package review

import "github.com/a-h/cap/model"

// buildChecklist returns the judgement questions appropriate to an entity kind. The
// questions encode the naming conventions and content schema
// as prompts for an external assessor.
func buildChecklist(kind model.Kind) []string {
	common := []string{
		"Is the name in sentence case (only the first word capitalised)?",
	}
	switch kind {
	case model.KindCapability:
		return append([]string{
			"Is the name an imperative verb plus noun, for example \"Evaluate policies\"?",
			"Does the name complete the sentence \"the context can ___\"?",
			"Is the name free of any user interface, technology, or implementation term?",
			"Is this capability distinct from every other capability in the model, not a duplicate under another name?",
			"Does the Description state the value or outcome delivered, not how it is implemented?",
			"Does the Scope make the boundary clear, with both in-scope and out-of-scope content?",
			"Will this capability survive UI redesigns, API changes, and technology migrations?",
			"Run the tests this capability lists under Verification, then check each invariant has a test that covers it. An invariant with no covering test is a quality gap.",
			"Does any test assert a behaviour that is not yet recorded as an invariant? If so, the invariant is missing and should be added.",
			"For each shared invariant this capability names, does that invariant also name this capability? A link declared on only one side is likely a half-deleted or forgotten relationship.",
		}, common...)
	case model.KindInvariant:
		return append([]string{
			"Is the invariant a rule that must always hold, stated as a positive assertion of behaviour?",
			"Does it state what must be true rather than how it is achieved?",
			"Is it a full sentence ending with a full stop?",
			"Is it verifiable, so that a specification or test could confirm it?",
			"For each capability this invariant names, does that capability also name this invariant? A link declared on only one side is likely a half-deleted or forgotten relationship.",
		}, common...)
	case model.KindSpecification:
		return append([]string{
			"Does the specification describe the design that realises the invariants: how the parts fit together, including implementation detail?",
			"Does it explain how, rather than restating the behaviour rules, which belong in the invariants?",
			"Does the design described match the capability or context this specification links to, rather than detailing a different one?",
		}, common...)
	case model.KindADR:
		return append([]string{
			"Does the ADR record the context, the decision, and the consequences?",
			"Are the alternatives that were considered, and why they were rejected, captured?",
			"Does the decision constrain how a capability is implemented, rather than describe the capability itself?",
		}, common...)
	case model.KindScenario:
		return append([]string{
			"Do the steps describe an end-to-end path a participant actually follows?",
			"Are the capabilities the scenario depends on all linked?",
			"Is the scenario named as the path, not as a capability?",
		}, common...)
	case model.KindVerification:
		return append([]string{
			"Run the tests. Do they pass?",
			"For each listed path, does the test name read as a positive assertion of behaviour, like an invariant?",
			"Map each test to the invariant it covers. A test with no matching invariant asserts a behaviour the model does not record; should that behaviour be added as an invariant?",
			"Do the listed paths exist and exercise the capabilities they verify?",
		}, common...)
	case model.KindTask:
		return append([]string{
			"Is the task named as a verb plus noun describing the work?",
			"Does the task clearly advance a capability?",
			"Is the task transient, rather than describing a long-lived capability?",
		}, common...)
	case model.KindContext:
		return append([]string{
			"Is the context a noun phrase naming a domain boundary, not a verb phrase?",
			"Do its capabilities genuinely belong together within one bounded context?",
		}, common...)
	}
	return common
}
