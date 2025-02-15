package yamlSchemaValidator

import (
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/datreeio/datree/pkg/fileReader"
	"github.com/datreeio/datree/pkg/policy"
	"gopkg.in/yaml.v2"
)

type TestFilesByRuleId = map[int]*FailAndPassTests

type FailAndPassTests struct {
	fails  []*FileWithName
	passes []*FileWithName
}

type FileWithName struct {
	name    string
	content string
}

func TestValidate(t *testing.T) {
	err := os.Chdir("../../")
	if err != nil {
		panic(err)
	}

	defaultRules, err := policy.GetDefaultRules()
	if err != nil {
		panic(err)
	}

	testFilesByRuleId := getTestFilesByRuleId(t)
	validator := New()

	for _, rule := range defaultRules.Rules {
		validatePassing(t, validator, rule.Schema, rule.ID, testFilesByRuleId[rule.ID].passes, true)
		validatePassing(t, validator, rule.Schema, rule.ID, testFilesByRuleId[rule.ID].fails, false)
	}
}

func validatePassing(t *testing.T, validator *YamlSchemaValidator, schemaContent map[string]interface{}, ruleId int, files []*FileWithName, expectPass bool) {
	for _, file := range files {
		schemaBytes, err := yaml.Marshal(schemaContent)
		if err != nil {
			panic(err)
		}

		res, err := validator.Validate(string(schemaBytes), file.content)
		if err != nil {
			panic(err)
		}

		if len(res.Errors()) > 0 && expectPass {
			t.Errorf("Expected validation for rule with id %d to pass, but it failed for file %s\n", ruleId, file.name)
		}
		if len(res.Errors()) == 0 && !expectPass {
			t.Errorf("Expected validation for rule with id %d to fail, but it passed for file %s\n", ruleId, file.name)
		}
	}
}

func getTestFilesByRuleId(t *testing.T) TestFilesByRuleId {
	dirPath := "./pkg/policy/tests"
	fileReader := fileReader.CreateFileReader(nil)
	files, err := fileReader.ReadDir(dirPath)
	if err != nil {
		panic(err)
	}

	testFilesByRuleId := make(TestFilesByRuleId)
	for _, file := range files {
		filename, err := fileReader.GetFilename(file)
		if err != nil {
			panic(err)
		}

		fileContent, err := fileReader.ReadFileContent(file)
		if err != nil {
			panic(err)
		}

		id, isPass := getFileData(filename)
		if testFilesByRuleId[id] == nil {
			testFilesByRuleId[id] = &FailAndPassTests{}
		}

		fileWithName := &FileWithName{name: filename, content: fileContent}
		if isPass {
			testFilesByRuleId[id].passes = append(testFilesByRuleId[id].passes, fileWithName)
		} else {
			testFilesByRuleId[id].fails = append(testFilesByRuleId[id].fails, fileWithName)
		}
	}

	return testFilesByRuleId
}

func getFileData(filename string) (int, bool) {
	parts := strings.Split(filename, "-")
	id, err := strconv.Atoi(parts[0])
	if err != nil {
		panic(err)
	}

	isPass := strings.Contains(parts[1], "pass")
	return id, isPass
}
