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
			"Does this capability do exactly one thing? A name built on \"Manage\", \"Handle\", \"Process\", or \"and\" usually bundles several capabilities and should be split into one per verb, for example \"Author policies\", \"Evaluate policies\", and \"Expire policies\".",
			"Is this capability at the right altitude: a single ability the system has, not a whole subsystem (too broad) and not one screen or endpoint (too narrow)?",
			"Is this capability distinct from every other capability in the model, not a duplicate under another name?",
			"Does the Description state the value or outcome delivered, not how it is implemented?",
			"Is this named and described by how the product is used, from the goal of the actor who uses it, rather than by an internal mechanism or data structure? A name like \"Bundle a capability context\" describes the plumbing that produces the output; \"Provide context to an agent\" describes what the actor gets and why. When the name echoes an internal type, function, or file, rename it to the use.",
			"Does the Description say why an actor reaches for this capability, so a reader learns what it is for and not only what it does?",
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
			"Does the design described match the capabilities or context this specification links to, rather than detailing a different one?",
			"For each capability this specification names, does that capability also name this specification? A link declared on only one side is likely a half-deleted or forgotten relationship.",
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
	case model.KindConcept:
		return append([]string{
			"Is the concept a domain noun, a thing the capabilities operate on, rather than an action or a rule?",
			"Does the Definition state what the thing is in the language of the domain, independently of how it is stored or represented in any technology?",
			"Does this concept belong to the context named in its Metadata, sharing that context's language?",
			"Is this concept distinct from every other concept, not the same thing under two names?",
			"Where this concept's name appears in another entity's text as a reference, is it capitalised and tagged with this identifier, for example \"Policy (con-0001)\", so the reference is marked and can be traced? Use sentence case for a multi-word name, for example \"Capability bundle (con-0006)\".",
		}, common...)
	case model.KindContext:
		return append([]string{
			"Is the context a noun phrase naming a domain boundary, not a verb phrase?",
			"Do its capabilities genuinely belong together within one bounded context?",
			"Is each thing this context operates on defined as a concept, rather than left implicit in capability names? The things are the nouns in its language, for example a policy, a decision, or a request.",
			"Does a noun recur across this context's entities without a concept defining it? A term used repeatedly with no definition, such as an undefined \"world state\", is a missing concept.",
			"For each concept, are all the verbs the system performs on it captured as capabilities? A concept that only ever appears as the object of one capability often has authoring, evaluation, and expiry capabilities missing.",
			"Are the capabilities here atomic, one verb on one concept, rather than a few broad capabilities like \"Manage policies\" that hide several finer-grained ones?",
			"Is each capability framed by how the product is used, from the goal of the actor who uses it, rather than by the mechanism that implements it? A capability named after an internal type or step, like \"Bundle a context\", should be renamed to the use it serves, like \"Provide context to an agent\".",
		}, common...)
	}
	return common
}
