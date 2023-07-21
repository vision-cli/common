package extractor

import (
	"bufio"
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

	"github.com/vision-cli/common/cases"
	"github.com/vision-cli/common/transpiler/model"
	protoparser "github.com/yoheimuta/go-protoparser"
	parserStructs "github.com/yoheimuta/go-protoparser/parser"
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

	modules := Modules(targetDir)

	for i, module := range modules {
		moduleNameWithVersion := fmt.Sprintf("%s.%s", module.Name, module.ApiVersion)
		moduleDirectory := filepath.Join(targetDir, moduleNameWithVersion)
		modules[i].Services = Services(moduleDirectory, &module)
	}

	//This is just used for testing to print the structure we're actually creating, for verification purposes
	for _, module := range modules {
		result += printFields(module, 1)
	}

	return result
}

func Modules(targetDir string) []model.Module {
	moduleDirs, _ := os.ReadDir(targetDir)
	modules := []model.Module{}

	for _, path := range moduleDirs {
		if path.IsDir() && path.Name() != "default" {
			parts := strings.Split(path.Name(), ".")
			if len(parts) == 2 {
				modules = append(modules, model.Module{
					Name:       parts[0],
					ApiVersion: parts[1],
				})
			} else {
				// Handle the case where the name doesn't contain a dot.
				// You may want to define a default value for ApiVersion here.
				modules = append(modules, model.Module{
					Name:       parts[0],
					ApiVersion: "", // Provide a default value or handle the case accordingly.
				})
			}
		}
	}

	return modules
}

func Services(moduleDirectory string, module *model.Module) []model.Service {
	serviceDirs, _ := os.ReadDir(moduleDirectory)
	services := []model.Service{}

	for i, path := range serviceDirs {
		if path.IsDir() {
			serviceName := path.Name()
			services = append(services, model.Service{Name: serviceName})

			// Construct the protoFilename
			protoFilename := fmt.Sprintf("%s_%s%s%s", module.Name, module.ApiVersion[:1], module.ApiVersion[1:], serviceName)
			protoFilename = cases.Snake(protoFilename)
			protoFilename = fmt.Sprintf("%s.proto", protoFilename)
			// Create the full protoFilePath
			protoFilePath := filepath.Join(moduleDirectory, serviceName, "proto", protoFilename)

			// Get the Enums for the service from the proto file
			services[i].Enums = getEnums(protoFilePath)

			// services[i].Enums = []model.Enum{{Name: "Test", Values: []string{"testing", "Enums"}}}

			modelsFolder := filepath.Join(moduleDirectory, path.Name(), "models", "models.go")
			services[i].Entities = getEntities(modelsFolder)
		}
	}

	return services
}

func getEntities(modelsGo string) []model.Entity {
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

func getEnums(protoFilePath string) []model.Enum {
	serviceEnums := []model.Enum{}

	// Open the file
	file, err := os.Open(protoFilePath)
	if err != nil {
		fmt.Println("Error opening file:", err)
	}
	defer file.Close()

	// Create an io.Reader to read from the file
	reader := bufio.NewReader(file)

	proto, _ := protoparser.Parse(reader)

	// Get enum information
	enumInfo := getEnumInformation(proto)

	// Print enum information
	for enumName, enumValues := range enumInfo {
		serviceEnums = append(serviceEnums, model.Enum{Name: enumName, Values: enumValues})
	}
	return serviceEnums
}

func getEnumInformation(proto *parserStructs.Proto) map[string][]string {
	enumInfo := make(map[string][]string)

	// Traverse ProtoBody and look for elements of type *Enum
	for _, element := range proto.ProtoBody {
		if enum, ok := element.(*parserStructs.Enum); ok {
			enumName := enum.EnumName
			enumValues := make([]string, len(enum.EnumBody))

			// Extract enum values from EnumBody
			for i, enumElement := range enum.EnumBody {
				if enumField, ok := enumElement.(*parserStructs.EnumField); ok {
					enumValues[i] = enumField.Ident
				}
			}

			// Add enum information to the map
			enumInfo[enumName] = enumValues
		}
	}

	return enumInfo
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
		case []string:
			result.WriteString("\n")
			for _, str := range val {
				result.WriteString(fmt.Sprintf("%s%s\n", indentStr, str))
			}
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
