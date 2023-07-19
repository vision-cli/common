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

	modules := getModules(targetDir)

	for i := range modules {
		moduleDirectory := filepath.Join(targetDir, modules[i].Name)
		modules[i].Services = getServices(moduleDirectory)
	}

	//This is just used for testing to print the structure we're actually creating, for verification purposes
	for _, module := range modules {
		result += printFields(module, 1)
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

	fset := token.NewFileSet()

	file, err := parser.ParseFile(fset, modelsGo, nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	for _, decl := range file.Decls {
		// Check if the declaration is a type declaration.
		if genDecl, ok := decl.(*ast.GenDecl); ok {
			parseTypeSpecs(genDecl.Specs, &entities)
		}
	}

	return entities
}

func parseTypeSpecs(specs []ast.Spec, entities *[]model.Entity) {
	for _, spec := range specs {
		if typeSpec, ok := spec.(*ast.TypeSpec); ok {
			if structType, ok := typeSpec.Type.(*ast.StructType); ok {
				if strings.HasSuffix(typeSpec.Name.Name, "Data") {
					entityName := typeSpec.Name.Name[:len(typeSpec.Name.Name)-4]
					fields := getFields(structType)

					entity := model.Entity{
						Name:        entityName,
						Persistence: "db",
						Fields:      fields,
					}

					*entities = append(*entities, entity)
				}
			}
		}
	}
}

func getFields(structType *ast.StructType) []model.Field {
	var fields []model.Field

	for _, field := range structType.Fields.List {
		if field.Names != nil {
			fields = append(fields, getFieldData(field))
		}
	}

	return fields
}

func getFieldData(field *ast.Field) model.Field {
	fieldData := model.Field{
		Name:         field.Names[len(field.Names)-1].Name,
		IsNullable:   false,
		IsSearchable: false,
	}

	switch fieldType := field.Type.(type) {
	case *ast.Ident:
		fieldData.Type = fieldType.Name
	case *ast.ArrayType:
		fieldData.Type = fieldType.Elt.(*ast.Ident).Name
		_, fieldData.IsArray = field.Type.(*ast.ArrayType)
	case *ast.SelectorExpr:
		fieldData.Type = getSelectorType(fieldType)
	default:
		fieldData.Type = "unknown"
	}

	return fieldData
}

func getSelectorType(selector *ast.SelectorExpr) string {
	selectorType := selector.Sel.Name
	if selectorType == "UUID" {
		return "id"
	}
	return "timestamp"
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
