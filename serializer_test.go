package viewset

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestDefaultSerializerSerialize(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	serializer := DefaultSerializer[testObject]{
		AdditionalField: map[string]Field[testObject]{
			"additions": testField{},
		},
	}
	object := testObject{Name: "test", Age: 18}
	result := gin.H{}

	serializer.Serialize(&result, &object, c)

	assert.Equal(t, object.Name, result["name"].(string))
	assert.Equal(t, object.Age, result["age"].(int))
	assert.Equal(t, "testing", result["additions"].(string))
}

func TestDefaultSerializerSerializeWithError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	serializer := DefaultSerializer[testObject]{
		AdditionalField: map[string]Field[testObject]{
			"additions": testField{},
		},
	}
	object := testObject{Name: "test", Age: 17}
	result := gin.H{}

	err := serializer.Serialize(&result, &object, c)

	assert.Equal(t, "Fail", err.Error())
}

func TestDefaultSerializerManySerialize(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	serializer := DefaultSerializer[testObject]{
		AdditionalField: map[string]Field[testObject]{
			"additions": testField{},
		},
	}
	objects := []*testObject{}
	objects = append(objects, &testObject{Name: "test", Age: 21}, &testObject{Name: "test 2", Age: 22})
	results := []gin.H{}

	serializer.ManySerialize(&results, &objects, c)

	assert.Equal(t, len(results), 2)
	for i, object := range objects {
		result := results[i]
		assert.Equal(t, object.Name, result["name"].(string))
		assert.Equal(t, object.Age, result["age"].(int))
		assert.Equal(t, "testing", result["additions"].(string))
	}
}

func TestDefaultSerializerManySerializeWithError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	serializer := DefaultSerializer[testObject]{
		AdditionalField: map[string]Field[testObject]{
			"additions": testField{},
		},
	}
	objects := []*testObject{}
	objects = append(objects, &testObject{Name: "test", Age: 21}, &testObject{Name: "test 2", Age: 17})
	results := []gin.H{}

	err := serializer.ManySerialize(&results, &objects, c)
	assert.Equal(t, "Fail", err.Error())
}
