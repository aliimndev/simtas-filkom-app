/ponytail-help

From an **application architecture perspective**, analyze this web application comprehensively and determine the most appropriate architectural approach for the current codebase.

**Do NOT modify or write any code yet. Perform the architectural analysis first.**

Base the analysis strictly on the **existing codebase**. Do not assume technologies, patterns, or structures that are not actually present.

### 1. Application Structure

* Is the current folder and module structure scalable?
* Is separation of concerns clear?
* Are there components, utilities, hooks, services, or business logic that are too tightly coupled?
* Identify duplicated responsibilities and misplaced logic.

### 2. Frontend Architecture

Evaluate:

* Server Components vs Client Components
* State management
* Data fetching and API communication
* Mutations
* Loading and error states
* Caching and revalidation
* UI components vs feature components
* Business logic placement
* Reusability and component boundaries

Determine whether the current boundaries are appropriate for a growing application.

### 3. Backend / API Architecture

Evaluate the separation between:

* Routes / Controllers
* Middleware
* Use Cases / Services
* Repositories
* Domain / Business Logic
* Database layer

Identify business logic that should not live inside controllers or HTTP handlers.

Also evaluate:

* Validation
* Error handling
* Authentication
* Authorization
* Middleware responsibilities
* Transaction boundaries

### 4. Database & Data Architecture

Analyze:

* Entity/model relationships
* Database constraints
* Transactions
* Data consistency
* Concurrency concerns
* Query efficiency
* Repository responsibilities
* Potential N+1 queries or inefficient access patterns

Business rules should not depend solely on frontend validation.

### 5. Authentication & Authorization

Review the complete authentication architecture, including:

* Login/session flow
* Token or cookie handling
* Refresh mechanism
* RBAC
* Permission boundaries
* Authentication middleware
* Authorization middleware

Identify architectural or security weaknesses.

### 6. Scalability & Maintainability

Determine whether the current architecture can comfortably support future feature growth.

Look for:

* Tight coupling
* Circular dependencies
* Excessive abstractions
* Duplicate logic
* God components
* God services
* Technical debt
* Unclear module boundaries

Avoid unnecessary enterprise patterns and over-engineering.

### 7. Performance Architecture

Identify potential bottlenecks related to:

* Rendering
* Client-side JavaScript
* API requests
* Database queries
* Asset loading
* Data fetching
* Caching

Only recommend optimizations that are justified by the actual codebase.

### 8. Recommended Architecture

After analyzing the current system, propose the architecture that best fits this project.

Explain:

* Recommended architectural pattern
* Recommended module/folder structure
* Dependency flow
* Layer boundaries
* Frontend → Backend communication
* Business logic placement
* Database access strategy
* External service integration boundaries

Clearly explain **why** the recommended architecture is better than the current approach.

### 9. Refactoring Strategy

Do NOT recommend a complete rewrite unless it is genuinely necessary.

Prefer incremental refactoring.

Categorize recommendations into:

**Priority 1 — Critical**
Issues that should be addressed immediately.

**Priority 2 — Important**
Issues that significantly improve maintainability, scalability, or reliability.

**Priority 3 — Improvement**
Nice-to-have architectural improvements that can be addressed later.

For every major recommendation, explain:

* Current problem
* Why it matters
* Recommended solution
* Expected benefit
* Risk of changing it

### 10. Final Architecture Assessment

Conclude with a concise architectural report containing:

* **Current Architecture**
* **Current Strengths**
* **Main Architectural Problems**
* **Recommended Architecture**
* **Recommended Module Boundaries**
* **Dependency Flow**
* **Critical Risks**
* **Refactoring Priorities**
* **What Should NOT Be Changed**

Use pragmatic engineering principles:

* KISS
* SOLID
* Separation of Concerns
* High Cohesion
* Low Coupling
* Dependency Inversion where appropriate
* YAGNI
* Pragmatic Architecture

Most importantly:

**Analyze the codebase that actually exists. Do not invent architecture, files, technologies, or problems that cannot be verified from the repository.**

At the end, provide a **clear architectural decision** rather than presenting many alternatives without choosing one.

Again: **ANALYZE ONLY. DO NOT MODIFY ANY FILES YET.**
