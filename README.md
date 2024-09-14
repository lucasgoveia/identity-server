# identity-server


/your_project
│
├── /cmd                  # Entry points for different commands (e.g., api, cli)
│   └── /api
│       └── main.go       # Main entry point for API server
│
├── /config               # Configuration files and environment setups
│
├── /pkg                  # Common/shared utilities, reusable across slices
│   └── /middleware       # Middleware functions (logging, authentication, etc.)
│   └── /logger           # Logger setup
│   └── /errors           # Global error handling
│
├── /internal             # Private application code
│   └── /app              # Application-level services, business rules
│   └── /domain           # Domain models, interfaces, and entities (shared across slices)
│
├── /features             # Vertical slices for each feature
│   └── /appointments     # Feature-specific slice
│       ├── /handlers     # HTTP handlers (controllers)
│       ├── /services     # Business logic
│       ├── /repository   # Data persistence layer
│       └── /dto          # Data Transfer Objects (e.g., request/response formats)
│   └── /invoices         # Another feature slice
│       ├── /handlers
│       ├── /services
│       ├── /repository
│       └── /dto
│
├── /migrations           # Database migrations
│
└── go.mod  