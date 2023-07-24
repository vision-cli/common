package extractor

import (
	"bufio"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/vision-cli/common/cases"
	"github.com/vision-cli/common/transpiler/model"
	protoparser "github.com/yoheimuta/go-protoparser"
	parserStructs "github.com/yoheimuta/go-protoparser/parser"
)

func ProjectStructure(projectDirectory string) ([]model.Module, error) {
	targetDir := filepath.Join(projectDirectory, "services")

	// Get module names and api versions
	modules, err := Modules(targetDir)
	if err != nil {
		return nil, fmt.Errorf("error reading modules: %w", err)
	}

	// Get services for each module
	for i, module := range modules {
		moduleNameWithVersion := fmt.Sprintf("%s.%s", module.Name, module.ApiVersion)
		moduleDirectory := filepath.Join(targetDir, moduleNameWithVersion)
		modules[i].Services, err = Services(moduleDirectory, &module)
		if err != nil {
			return nil, fmt.Errorf("error reading services: %w", err)
		}
	}

	return modules, nil
}

func Modules(targetDir string) ([]model.Module, error) {
	moduleDirs, err := os.ReadDir(targetDir)
	if err != nil {
		return nil, fmt.Errorf("error reading directory, please ensure project root is correct: %w", err)
	}

	modules := make([]model.Module, 0, len(moduleDirs))

	for _, path := range moduleDirs {
		if path.IsDir() && strings.Contains(path.Name(), ".") {
			parts := strings.Split(path.Name(), ".")
			if len(parts) == 2 {
				modules = append(modules, model.Module{
					Name:       parts[0],
					ApiVersion: parts[1],
				})
			}
		}
	}

	return modules, nil
}

func Services(moduleDirectory string, module *model.Module) ([]model.Service, error) {
	serviceDirs, err := os.ReadDir(moduleDirectory)
	if err != nil {
		return nil, fmt.Errorf("error reading module directory: %w", err)
	}

	services := []model.Service{}

	for i, path := range serviceDirs {
		if path.IsDir() {
			serviceName := path.Name()
			services = append(services, model.Service{Name: serviceName})

			// Construct the .proto file path
			protoFilename := fmt.Sprintf("%s_%s%s%s", module.Name, module.ApiVersion[:1], module.ApiVersion[1:], serviceName)
			protoFilename = cases.Snake(protoFilename)
			protoFilename = fmt.Sprintf("%s.proto", protoFilename)
			protoFilePath := filepath.Join(moduleDirectory, serviceName, "proto", protoFilename)

			services[i].Enums, err = getEnums(protoFilePath)
			if err != nil {
				return nil, fmt.Errorf("error getting enums: %w", err)
			}

			modelsFolder := filepath.Join(moduleDirectory, path.Name(), "models", "models.go")
			services[i].Entities, err = getEntities(modelsFolder)
			if err != nil {
				return nil, fmt.Errorf("error getting entities: %w", err)
			}
		}
	}

	return services, nil
}

func getEnums(protoFilePath string) ([]model.Enum, error) {
	serviceEnums := []model.Enum{}

	file, err := os.Open(protoFilePath)
	if err != nil {
		return nil, fmt.Errorf("error opening .proto file: %w", err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	proto, _ := protoparser.Parse(reader)

	enumInfo := getEnumInformation(proto)

	for enumName, enumValues := range enumInfo {
		serviceEnums = append(serviceEnums, model.Enum{Name: enumName, Values: enumValues})
	}
	return serviceEnums, nil
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

			enumInfo[enumName] = enumValues
		}
	}

	return enumInfo
}

func getEntities(modelsGo string) ([]model.Entity, error) {
	entities := []model.Entity{}

	fset := token.NewFileSet()

	file, err := parser.ParseFile(fset, modelsGo, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("error parsing models.go: %w", err)
	}

	// Default to in memory
	persistence := "memory"

	for _, decl := range file.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok {
			parseTypeSpecs(genDecl.Specs, &entities, persistence)

			persistence = getPersistence(genDecl.Specs)
		}
	}

	return entities, nil
}

func getPersistence(specs []ast.Spec) string {
	for _, spec := range specs {
		if typeSpec, ok := spec.(*ast.TypeSpec); ok {
			if structType, ok := typeSpec.Type.(*ast.StructType); ok {
				if usesGorm(structType) {
					return "db"
				}
			}
		}
	}

	// Default to memory
	return "memory"
}

func parseTypeSpecs(specs []ast.Spec, entities *[]model.Entity, persistence string) {

	for _, spec := range specs {
		if typeSpec, ok := spec.(*ast.TypeSpec); ok {
			if structType, ok := typeSpec.Type.(*ast.StructType); ok {
				if strings.HasSuffix(typeSpec.Name.Name, "Data") {
					entityName := typeSpec.Name.Name[:len(typeSpec.Name.Name)-4]
					fields := getFields(structType)

					entity := model.Entity{
						Name:        entityName,
						Persistence: persistence,
						Fields:      fields,
					}

					*entities = append(*entities, entity)
				}
			}
		}
	}
}

func usesGorm(structType *ast.StructType) bool {
	for _, field := range structType.Fields.List {

		if field.Tag != nil {
			// Assuming the field tag is a string, so remove backticks and split by spaces.
			tagValues := strings.Fields(strings.Trim(field.Tag.Value, "`"))
			for _, tagValue := range tagValues {
				if strings.HasPrefix(tagValue, "gorm:") {
					// We found a gorm tag.
					return true
				}
			}
		}
	}
	return false
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

	if selectorType == "gorm" {
		return "gorm"
	}

	return "timestamp"
}
