package pathenvconfig

import (
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFieldNameToVariable(t *testing.T) {
	cases := []struct {
		fieldName string
		exected   string
	}{
		{"D", "PREFIX_D"},
		{"DB", "PREFIX_DB"},
		{"IOS", "PREFIX_IOS"},
		{"IOS1", "PREFIX_IOS_1"},
		{"IOS10", "PREFIX_IOS_10"},
		{"IAm", "PREFIX_I_AM"},
		{"Database", "PREFIX_DATABASE"},
		{"Database1", "PREFIX_DATABASE1"},
		{"Database10", "PREFIX_DATABASE10"},
		{"Database_10", "PREFIX_DATABASE_10"},
		{"DatabaseName", "PREFIX_DATABASE_NAME"},
		{"DatabaseConnectionString", "PREFIX_DATABASE_CONNECTION_STRING"},
		{"SSLCert", "PREFIX_SSL_CERT"},
		{"d", "PREFIX_D"},
		{"database", "PREFIX_DATABASE"},
		{"databaseCert", "PREFIX_DATABASE_CERT"},
		{"database_cert", "PREFIX_DATABASE_CERT"},
		{"database_cert", "PREFIX_DATABASE_CERT"},
		{"database_10", "PREFIX_DATABASE_10"},
	}

	for _, c := range cases {

		t.Run(c.fieldName, func(t *testing.T) {
			assert.Equal(t, c.exected, fieldNameToEnvironmentVar("PREFIX_", c.fieldName))
		})
	}
}

type ConfigSpec struct {
	Name   string `required:"true"`
	Age    int    `default:"13"`
	IsDog  bool
	Weight *int
}

func TestEnvionmentVariables(t *testing.T) {
	setVariable(t.Name(), "NAME", "Oona")
	setVariable(t.Name(), "AGE", "3")
	setVariable(t.Name(), "IS_DOG", "true")
	setVariable(t.Name(), "WEIGHT", "50")

	spec := ConfigSpec{}
	require.Nil(t, Process(t.Name(), &spec))
	assert.Equal(t, "Oona", spec.Name)
	assert.Equal(t, 3, spec.Age)
	assert.True(t, spec.IsDog)
	assert.Equal(t, 50, *spec.Weight)
}

func TestEnvionmentVariablesNoPrefix(t *testing.T) {
	os.Setenv("NAME", "Charlie")
	defer os.Unsetenv("NAME")
	os.Setenv("AGE", "8")
	defer os.Unsetenv("AGE")
	os.Setenv("IS_DOG", "true")
	defer os.Unsetenv("IS_DOG")

	spec := ConfigSpec{}
	require.Nil(t, Process("", &spec))
	assert.Equal(t, "Charlie", spec.Name)
	assert.Equal(t, 8, spec.Age)
	assert.True(t, spec.IsDog)
}

func TestDefaults(t *testing.T) {
	setVariable(t.Name(), "NAME", "Tom")
	spec := ConfigSpec{}
	require.Nil(t, Process(t.Name(), &spec))
	assert.Equal(t, "Tom", spec.Name)
	assert.Equal(t, 13, spec.Age)
	assert.False(t, spec.IsDog)
}

func TestRequired(t *testing.T) {
	spec := ConfigSpec{}
	require.NotNil(t, Process(t.Name(), &spec))
}

func TestFileEnvionmentVariables(t *testing.T) {
	namePath := path.Join(t.TempDir(), "name.txt")
	require.Nil(t, os.WriteFile(namePath, []byte("Alice"), 0644))
	setVariable(t.Name(), "NAME_FILE", namePath)

	agePath := path.Join(t.TempDir(), "age.txt")
	require.Nil(t, os.WriteFile(agePath, []byte("10"), 0644))
	setVariable(t.Name(), "AGE_FILE", agePath)

	isDogPath := path.Join(t.TempDir(), "isdog.txt")
	require.Nil(t, os.WriteFile(isDogPath, []byte("true"), 0644))
	setVariable(t.Name(), "IS_DOG_FILE", isDogPath)

	spec := ConfigSpec{}
	require.Nil(t, Process(t.Name(), &spec))
	assert.Equal(t, "Alice", spec.Name)
	assert.Equal(t, 10, spec.Age)
	assert.True(t, spec.IsDog)
}

func TestValueWithSpaces(t *testing.T) {
	setVariable(t.Name(), "NAME", "Oona the Dog")

	spec := ConfigSpec{}
	require.Nil(t, Process(t.Name(), &spec))
	assert.Equal(t, "Oona the Dog", spec.Name)
}

func TestTailingNewlineRemovedFromFile(t *testing.T) {
	// verify that the trailing newline is removed when reading from a file
	testCases := []struct {
		input    string
		expected string
	}{
		{"Ralph", "Ralph"},
		{"Ralph\n", "Ralph"},
		{"Ralph\r\n", "Ralph"},
		{"Ralph\n", "Ralph"},
		{"\nRalph\n", "\nRalph"},
	}
	for _, tC := range testCases {
		t.Run(tC.input, func(t *testing.T) {
			namePath := path.Join(t.TempDir(), "name.txt")
			require.Nil(t, os.WriteFile(namePath, []byte(tC.input), 0644))
			setVariable(t.Name(), "NAME_FILE", namePath)
			spec := ConfigSpec{}
			require.Nil(t, Process(t.Name(), &spec))
			assert.Equal(t, tC.expected, spec.Name)
		})
	}

	// But not removed from environment variables.
	setVariable(t.Name(), "NAME", "Oona\n")
	spec := ConfigSpec{}
	require.Nil(t, Process(t.Name(), &spec))
	assert.Equal(t, "Oona\n", spec.Name)
}

func setVariable(prefix, name, value string) {
	os.Setenv(fmt.Sprintf("%s_%s", prefix, name), value)
}

type DatabaseConfig struct {
	User     string
	Password string
}

type WithNesteStruct struct {
	Db     DatabaseConfig
	DbPtr  *DatabaseConfig
	DbPtr2 *DatabaseConfig
	DbPtr3 *DatabaseConfig
}

func TestNestedMap(t *testing.T) {
	setVariable(t.Name(), "DB_USER", "Oscar")
	setVariable(t.Name(), "DB_PTR_USER", "Bert")

	spec := WithNesteStruct{DbPtr3: &DatabaseConfig{}}
	require.Nil(t, Process(t.Name(), &spec))
	assert.Equal(t, "Oscar", spec.Db.User)
	assert.Equal(t, "Bert", spec.DbPtr.User)
	assert.Nil(t, spec.DbPtr2)
	assert.NotNil(t, spec.DbPtr3)
}
