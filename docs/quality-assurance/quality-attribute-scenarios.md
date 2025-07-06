# Quality Attribute Scenarios

## Functional Suitability
### Functional Correctness

#### General description
Functional correctness is important for the customer as it directly influences on 
the quality of the AI answers and, correspondingly, convenience and usefulness of
the consultant. If the answers are incorrect, inaccurate, or meaningless, this will:

* directly affect the trust of customers,
* lead to a loss of sales,
* create reputation risks (false recommendations, price errors, etc.).

AI always must provide relevant and accurate answers.

#### QAS-tests
##### Correct answer for specific product request
- **Source**: Customer
- **Stimulus**: Asks "Which laptops are suitable for gaming up to $1,000?"
- **Artifact**: AI question-answering module
- **Environment**: Normal load, production
- **Response**: Returns a list of gaming laptops under $1000
- **Response measure**: The result includes only laptops matching the criteria, verified against product database

##### Correct handling of a question that must not be answered
- **Source of stimulus**: Customer
- **Stimulus**: Asks "Tell me about baking bread"
- **Artifact**: AI question-answering module
- **Environment**: Normal load, production
- **Response**: Gracefully informs that the consultant cannot answer this question
- **Response measure**: System gives a negative but polite answer without humiliating the honor and dignity of the user
##### Method of testing
Conducting unit, integration and user acceptance tests that compare actual responses with reference ones.

## Interaction Capability
### User assistance
#### General description
Users of the online store can ask the AI consultant unstructured, incomplete, or 
ambiguous questions. In order not to lose customers, the system must prompt, 
clarify and help formulate requests, not say "error" and/or remain the user without answer.

User assistance makes the AI consultant friendly and convenient, 
even for those who do not formulate questions correctly.
#### QAS-tests
##### Assist when input is empty or vague
- **Source of stimulus**: User
- **Stimulus**: Sends an empty or vague request (e.g. "something", "help me")
- **Artifact**: AI question-answering module
- **Environment**: Web interface, normal conditions
- **Response**: System returns a friendly clarification or prompt
- **Response measure**: Response contains a hint for better formulation
- od the question and clarifying what does the user need.

##### Method of testing
Send a request with an empty or unclear question to the AI.
Check that the response contains a clarification prompt.
## Maintainability
### Modularity 
#### General description
Online store (and, correspondingly, the AI-assistant) always can be modified:
new categories of good can be added, new model of AI can be used for the answers, new functionality
can be added. If the online-consultant lacks the modularity, any changes are:
* expensive,
* risky,
* long in time

to avoid crashes of the consultant. It is important to the customer that his 
team can easily replace some part of the assistant (e. g., an ML model) without 
interfering with the rest of the system.
#### QAS-tests
##### Change product data format without breaking chatbot
- **Source of stimulus**: Backend engineer
- **Stimulus**: Updates schema of product database (e.g., remove the registration feature)
- **Artifact**: Product data parser
- **Environment**: Test environment
- **Response**: Chatbot still returns product recommendations
- **Response measure**: Product-related queries pass test suite without chatbot 
code change
##### Method of testing
Conducting unit, integration and user acceptance tests for checking the stable work of the
rest code blocks.   
