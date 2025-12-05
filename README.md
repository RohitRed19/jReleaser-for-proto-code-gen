# jReleaser for Proto Code Gen

A multi-language Protocol Buffer code generation and release management project using Maven. This project demonstrates how to generate gRPC/protobuf code for **Java**, **Go**, **Python**, and **Rust** from a single `.proto` definition, and deploy the generated artifacts to a local Nexus repository.

## Features

- **Multi-language protobuf compilation**: Generate gRPC stubs and protobuf messages for Java, Go, Python, and Rust
- **Centralized version management**: Maven manages versions across all language targets
- **Local Nexus deployment**: Deploy Maven JARs, Python packages (PyPI), Go modules, and Rust crates to a local Nexus repository
- **Automated Go module versioning**: Go modules are automatically versioned based on Maven project version

## Project Structure

```
.
├── pom.xml                    # Parent POM with shared configuration
├── manage.go                  # Go utility for module version management
├── hello/                     # Example service module
│   ├── pom.xml
│   ├── src/main/protobuf/     # Proto definitions
│   │   └── hello.proto
│   └── rust/                  # Rust crate (optional)
│       ├── Cargo.toml
│       └── src/
├── services-parent/           # Parent POM for service modules
│   ├── pom.xml                # Plugin configurations for all languages
│   ├── go/                    # Generated Go code
│   └── setup.py.template      # Python package template
└── settings.xml               # Maven settings for Nexus integration
```

## Prerequisites

### SDKMAN! (Java & Maven)

This project uses [SDKMAN!](https://sdkman.io/) for Java and Maven version management. Install SDKMAN:

```bash
curl -s "https://get.sdkman.io" | bash
```

Then install the required versions:

```bash
sdk env install
```

This will install the versions specified in `.sdkmanrc`:
- Java 21 (Temurin)
- Maven 3.9.9

### Aqua (Go & Protoc Plugins)

[Aqua](https://aquaproj.github.io/) manages Go and protobuf code generators.

#### Install Aqua

```bash
# macOS (Homebrew)
brew install aquaproj/aqua/aqua

# Linux / macOS (Shell script)
curl -sSfL https://raw.githubusercontent.com/aquaproj/aqua-installer/v3.1.1/aqua-installer | bash

# Windows (Scoop)
scoop install aqua
```

#### Install Go Tools

```bash
aqua install
```

This installs tools defined in `aqua.yaml`:

| Tool | Version | Purpose |
|------|---------|---------|
| `go` | 1.22.5 | Go runtime |
| `protoc-gen-go` | 1.36.1 | Go protobuf code generator |
| `protoc-gen-go-grpc` | 1.5.1 | Go gRPC code generator |

> **Note**: The `protoc` compiler is automatically downloaded by the Maven protobuf plugin during build.

### Rust (Optional)

For Rust code generation, install [Rust](https://rustup.rs/):

```bash
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
rustup default stable
```

### Python

Python 3 with pip is required for Python code generation:

```bash
pip install grpcio-tools build twine
```

### Local Nexus Repository

A local Nexus Repository Manager is required for artifact deployment. See [nexus-data/READMe.md](nexus-data/READMe.md) for setup instructions.

## Usage

### Build All Artifacts

```bash
mvn clean compile
```

This will:
1. Generate Java protobuf/gRPC code
2. Generate Go protobuf/gRPC code (versioned modules)
3. Generate Python protobuf/gRPC code
4. Compile Rust crate (if present)

### Package Artifacts

```bash
mvn package
```

Creates:
- Java JAR files
- Python wheel and source distributions

### Deploy to Nexus

```bash
export NEXUS_USER=admin
export NEXUS_PASSWORD=your-password
mvn deploy -s settings.xml
```

Deploys:
- Java JARs to Maven repository
- Python packages to PyPI repository
- Rust crates to Cargo registry (if present)

### Release a New Version

```bash
mvn release:prepare release:perform -s settings.xml
```

## Configuration

### Maven Settings

Copy `settings.xml` to `~/.m2/settings.xml` or use `-s settings.xml` flag:

```xml
<servers>
    <server>
        <id>local-nexus</id>
        <username>${NEXUS_USER}</username>
        <password>${NEXUS_PASSWORD}</password>
    </server>
</servers>
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `NEXUS_USER` | Nexus repository username |
| `NEXUS_PASSWORD` | Nexus repository password |

## Adding a New Service

1. Create a new directory under the project root:

```bash
mkdir -p my-service/src/main/protobuf
```

2. Add your `.proto` file(s) to `my-service/src/main/protobuf/`

3. Create `my-service/pom.xml`:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0"
         xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 
                             http://maven.apache.org/xsd/maven-4.0.0.xsd">
    <modelVersion>4.0.0</modelVersion>
    <parent>
        <groupId>com.example.jreleaser</groupId>
        <artifactId>services-parent</artifactId>
        <version>1.0.0-SNAPSHOT</version>
        <relativePath>../services-parent</relativePath>
    </parent>

    <artifactId>my-service</artifactId>

    <build>
        <plugins>
            <plugin>
                <groupId>io.github.ascopes</groupId>
                <artifactId>protobuf-maven-plugin</artifactId>
            </plugin>
            <plugin>
                <groupId>org.codehaus.mojo</groupId>
                <artifactId>exec-maven-plugin</artifactId>
            </plugin>
            <plugin>
                <groupId>org.codehaus.mojo</groupId>
                <artifactId>build-helper-maven-plugin</artifactId>
            </plugin>
        </plugins>
    </build>
</project>
```

4. Add the module to the root `pom.xml`:

```xml
<modules>
    <module>hello</module>
    <module>my-service</module>
    <module>services-parent</module>
</modules>
```

## Go Module Versioning

The `manage.go` utility handles Go module versioning automatically:

- Extracts major version from Maven `pom.xml`
- Creates versioned directories (`go/v1`, `go/v2`, etc.)
- Updates `go.mod` module paths for major version changes

Run manually if needed:

```bash
go run manage.go upgrade-mod services-parent
go run manage.go clean services-parent/go
```

## Dependencies

### Java
- Protobuf 4.33.1
- gRPC 1.77.0
- Guava 33.5.0-jre

### Rust
- tonic 0.14.2
- prost 0.14.1
- tokio 1.48.0

## License

MIT License - see [LICENSE](LICENSE) for details.
