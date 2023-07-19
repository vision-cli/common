package extractor

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/vision-cli/common/transpiler/model"
)

// type Module struct {
// 	ApiVersion string    `yaml:"apiVersion"`
// 	Name       string    `yaml:"name"`
// 	Services   []Service `yaml:"services"`
// }

// service := model.Service{
// 	Name: "projects",
// 	Enums: []model.Enum{
// 		{Name: "project-type", Values: []string{"not-assigned", "internal", "billable"}},
// 	},
// 	Entities: []model.Entity{
// 		{
// 			Name:        "project",
// 			Persistence: "db",
// 			Fields: []model.Field{
// 				{Name: "name", Type: "string", Tag: "db:", IsArray: false, IsNullable: true, IsSearchable: false},
// 			},
// 		},
// 	},
// }

func GetProjectStructure(projectDirectory string) string {
	var result string
	targetDir := filepath.Join(projectDirectory, "services")

	//Getting Module Names
	modules := getModules(targetDir)

	//Getting services names
	for i := range modules {
		moduleDirectory := filepath.Join(targetDir, modules[i].Name)
		modules[i].Services = getServices(moduleDirectory)
	}

	//This is just used for testing to print the structure we're actually creating, for verification purposes
	for _, module := range modules {
		result += printFields(module, 1)
		// result += fmt.Sprintf("\nModule Name is: "+module.Name+" - Module index is: %d\n", i)
		// for j, service := range module.Services {
		// 	result += fmt.Sprintf("\tService Name is: "+service.Name+" - Service index is: %d\n", j)
		// 	for k, entity := range service.Entities {
		// 		result += fmt.Sprintf("\t\tEntity Name is: "+entity.Name+" - Entity index is: %d\n", k)
		// 		for l, field := range entity.Fields {
		// 			result += fmt.Sprintf("\t\t\tField Value is: "+field.Name+" - Field Value index is: %d\n", l)
		// 		}
		// 	}
		// }
	}

	return "[]model.Module{}, []model.Service{}sameas: " + result
}

func getModules(targetDir string) []model.Module {
	moduleDirs, _ := os.ReadDir(targetDir)
	modules := []model.Module{}

	for _, path := range moduleDirs {
		if path.IsDir() && path.Name() != "default" {
			modules = append(modules, model.Module{Name: path.Name()})
		}
	}

	return modules
}

// service := model.Service{
// 	Name: "projects",
// 	Enums: []model.Enum{
// 		{Name: "project-type", Values: []string{"not-assigned", "internal", "billable"}},
// 	},
// 	Entities: []model.Entity{
// 		{
// 			Name:        "project",
// 			Persistence: "db",
// 			Fields: []model.Field{
// 				{Name: "name", Type: "string", Tag: "db:", IsArray: false, IsNullable: true, IsSearchable: false},
// 			},
// 		},
// 	},
// }

func getServices(moduleDirectory string) []model.Service {
	serviceDirs, _ := os.ReadDir(moduleDirectory)
	services := []model.Service{}

	for i, path := range serviceDirs {
		if path.IsDir() {
			services = append(services, model.Service{Name: path.Name()})
			modelsFolder := filepath.Join(moduleDirectory, path.Name(), "models")
			// services[i].Enums = getEnums(modelsFolder)
			// services[i].Enums = []model.Enum{{Name: "Test", Values: []string{"testing", "Enums"}}}
			services[i].Entities = getEntities(modelsFolder)
		}
	}

	return services
}

func getEntities(modelsFolder string) []model.Entity {
	modelsGo := filepath.Join(modelsFolder, "models.go")
	entities := []model.Entity{}

	// entities = append(entities, model.Entity{Name: modelsGo, Persistence: "db", Fields: []model.Field{{Name: "FieldName"}}})

	// Create a new file set
	fset := token.NewFileSet()

	// Parse the file and retrieve the AST
	file, err := parser.ParseFile(fset, modelsGo, nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	// Loop through the declarations in the file.
	// for _, decl := range file.Decls {
	// 	// Check if the declaration is a type declaration.
	// 	if genDecl, ok := decl.(*ast.GenDecl); ok {
	// 		// Check if the type declaration has specifications (e.g., structs).
	// 		for _, spec := range genDecl.Specs {
	// 			if typeSpec, ok := spec.(*ast.TypeSpec); ok {
	// 				// Check if the specification is a struct type.
	// 				if _, ok := typeSpec.Type.(*ast.StructType); ok {
	// 					// Found the first struct type. Print its name and exit.
	// 					if !strings.HasSuffix(typeSpec.Name.Name, "Data") {
	// 						//Get FIELDS and pass them to entities
	// 						var fields []model.Field
	// 						entities = append(entities, model.Entity{Name: typeSpec.Name.Name, Persistence: "db", Fields: fields})
	// 					}
	// 				}
	// 			}
	// 		}
	// 	}
	// }

	for _, decl := range file.Decls {
		// Check if the declaration is a type declaration.
		if genDecl, ok := decl.(*ast.GenDecl); ok {
			for _, spec := range genDecl.Specs {
				if typeSpec, ok := spec.(*ast.TypeSpec); ok {
					// Check if the specification is a struct type.
					if structType, ok := typeSpec.Type.(*ast.StructType); ok {
						// Check if the struct type is named "ProjectData".
						if strings.HasSuffix(typeSpec.Name.Name, "Data") {
							// Create an Entity for "Project" and add fields from "ProjectData".
							entityName := typeSpec.Name.Name[:len(typeSpec.Name.Name)-4]
							entity := model.Entity{Name: entityName}

							fields := getFields(structType)
							// Loop through the fields of ProjectData.

							entity.Fields = fields

							// Add the entity to the entities slice.
							entities = append(entities, entity)
						}
					}
				}
			}
		}
	}
	// Process the AST as needed
	// Example: Print the names of all struct types
	// ast.Inspect(file, func(node ast.Node) bool {
	// 	if typeSpec, ok := node.(*ast.TypeSpec); ok {
	// 		if structType, ok := typeSpec.Type.(*ast.StructType); ok {
	// 			entities[0].Name = fmt.Sprint("Struct Name:", typeSpec.Name)
	// 			for _, field := range structType.Fields.List {
	// 				entities[0].Persistence = fmt.Sprint("Field Name:", field.Names)
	// 			}
	// 		}
	// 	}
	// 	return true
	// })

	return entities
}

func getFields(structType *ast.StructType) []model.Field {
	// Loop through the fields of ProjectData.
	var fields []model.Field

	for _, field := range structType.Fields.List {
		if field.Names != nil {
			// Field names can be different if there are multiple fields in one line.
			// In this case, we take only the last field name.
			name := field.Names[len(field.Names)-1].Name

			// Extract type information from the field.
			var fieldType string
			if _, isIdent := field.Type.(*ast.Ident); isIdent {
				fieldType = field.Tag.Kind.String()
			} else if selector, isSelector := field.Type.(*ast.SelectorExpr); isSelector {
				selectorType := selector.Sel.Name
				if selectorType == "UUID" {
					fieldType = "id"
				} else {
					fieldType = "timestamp"
				}
			}
			// You can process fieldType to get the actual type information.
			// For example, if it's a pointer type, array, or a custom type, etc.

			// Determine if the field is an array or not.
			_, isArray := field.Type.(*ast.ArrayType)

			// if arrType, isArray := field.Type.(*ast.ArrayType); isArray {
			// 	// Found an array type.
			// 	// Do whatever processing you need with the array type here.
			// 	fmt.Println("Field:", field.Names[0].Name)
			// 	fmt.Println("Type is an array:", arrType.Len == nil) // Check if the array type has a length or is a slice.
			// }
			// You can inspect fieldType further to determine if it's an array type.

			// Determine if the field is nullable (i.e., a pointer type).
			isNullable := false
			// You can check if fieldType is a pointer type.

			// Determine if the field is searchable (if applicable in your context).
			isSearchable := false
			// You can define criteria for determining if the field is searchable.

			// Append the field to the entity.Fields slice.
			fields = append(fields, model.Field{
				Name:         name,
				Type:         fieldType,
				IsArray:      isArray,
				IsNullable:   isNullable,
				IsSearchable: isSearchable,
			})
		}
	}

	return fields
}

func getEnums(modelsFolder string) []model.Enum {
	modelsGo := filepath.Join(modelsFolder, "models.go")
	enums := []model.Enum{}

	enums = append(enums, model.Enum{Name: modelsGo, Values: []string{"Temps"}})

	// Create a new file set
	// fset := token.NewFileSet()

	// Parse the file and retrieve the AST
	// file, err := parser.ParseFile(fset, modelsGo, nil, parser.ParseComments)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// Process the AST as needed
	// Example: Print the names of all struct types
	// ast.Inspect(file, func(node ast.Node) bool {
	// 	if typeSpec, ok := node.(*ast.TypeSpec); ok {
	// 		if structType, ok := typeSpec.Type.(*ast.StructType); ok {
	// 			enums[0].Values[0] = fmt.Sprint("Struct Name:", typeSpec.Name)
	// 			for _, field := range structType.Fields.List {
	// 				enums[0].Values[1] = fmt.Sprint("Field Name:", field.Names)
	// 			}
	// 		}
	// 	}
	// 	return true
	// })

	return enums
}

func printFields(data interface{}, indent int) string {
	v := reflect.ValueOf(data)
	t := reflect.TypeOf(data)

	var result strings.Builder

	for i := 0; i < v.NumField(); i++ {
		fieldName := t.Field(i).Name
		fieldValue := v.Field(i).Interface()

		indentStr := strings.Repeat(" ", indent)

		result.WriteString(fmt.Sprintf("%s%s: ", indentStr, fieldName))

		switch val := fieldValue.(type) {
		case string:
			result.WriteString(fmt.Sprintf("%s\n", val))
		case []model.Service:
			result.WriteString("\n")
			for _, service := range val {
				result.WriteString(printFields(service, indent+1))
			}
		case []model.Entity:
			result.WriteString("\n")
			for _, entity := range val {
				result.WriteString(printFields(entity, indent+2))
			}
		case []model.Enum:
			result.WriteString("\n")
			for _, enum := range val {
				result.WriteString(printFields(enum, indent+3))
			}
		case []model.Field:
			result.WriteString("\n")
			for _, field := range val {
				result.WriteString(printFields(field, indent+3))
			}
		case bool:
			result.WriteString(strconv.FormatBool(val))
			result.WriteString("\n")
		default:
			result.WriteString(fmt.Sprintf("Unhandled type: %s\n", reflect.TypeOf(val)))
		}
	}

	return result.String()
}
