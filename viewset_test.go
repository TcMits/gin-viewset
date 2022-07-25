package viewset

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestNewViewSet(t *testing.T) {
	basePath := "/objects"
	additionalActions := []Route[testObject, testObjectRequest]{}
	additionalActions = append(additionalActions,
		Route[testObject, testObjectRequest]{
			Action:  "send",
			SubPath: "/:pk/send",
			Method:  http.MethodPut,
			Handler: func(_ string, _ *ViewSet[testObject, testObjectRequest], _ *gin.Context) {
			},
		},
	)

	objectManager := &testObjectManager{}
	viewSet := NewViewSet[testObject, testObjectRequest](
		basePath, "/:pk", nil, additionalActions, objectManager, nil, nil, nil, nil,
	)

	assert.Equal(t, basePath, viewSet.BasePath)
	assert.Equal(t, 7, len(viewSet.Actions))
	assert.Equal(t, objectManager, viewSet.Manager)
	assert.NotEqual(t, nil, viewSet.ExceptionHandler)
	assert.NotEqual(t, nil, viewSet.PermissionChecker)
	assert.NotEqual(t, nil, viewSet.Serializer)
	assert.NotEqual(t, nil, viewSet.FormValidator)
}

func TestNewViewSetWithExcludeActions(t *testing.T) {
	basePath := "/objects"
	excludeActions := []string{}
	excludeActions = append(excludeActions, DEFAULT_UPDATE_ACTION)

	objectManager := &testObjectManager{}
	viewSet := NewViewSet[testObject, testObjectRequest](
		basePath, "/:pk", excludeActions, nil, objectManager, nil, nil, nil, nil,
	)

	assert.Equal(t, basePath, viewSet.BasePath)
	assert.Equal(t, 4, len(viewSet.Actions))
}

func TestViewSetRegister(t *testing.T) {
	basePath := "/objects"

	objectManager := &testObjectManager{}
	viewSet := NewViewSet[testObject, testObjectRequest](
		basePath, "/:pk", nil, nil, objectManager, nil, nil, nil, nil,
	)
	router := SetUpRouter()
	viewSet.Register(router)

	assert.Equal(t, 6, len(router.Routes()))
}

func TestGetHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	basePath := "/objects"
	objectManager := &testObjectManager{}
	viewSet := NewViewSet[testObject, testObjectRequest](
		basePath, "/:pk", nil, nil, objectManager, nil, nil, nil, nil,
	)
	callBack := ""

	handler := getHandler(
		"test",
		*viewSet,
		func(s string, vs *ViewSet[testObject, testObjectRequest], ctx *gin.Context) {
			callBack = "test"
		},
	)

	handler(c)
	assert.Equal(t, "test", callBack)
}

func TestGetHandlerWithPermissionError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	blw := &MockGinBodyResponseWriter{MockBody: bytes.NewBufferString(""), ResponseWriter: c.Writer}
	c.Writer = blw

	basePath := "/objects"
	permissionChecker := &MockDeniedAny{}

	objectManager := &testObjectManager{}
	viewSet := NewViewSet[testObject, testObjectRequest](
		basePath, "/:pk", nil, nil, objectManager, nil, permissionChecker, nil, nil,
	)

	handler := getHandler(
		"test",
		*viewSet,
		func(s string, vs *ViewSet[testObject, testObjectRequest], ctx *gin.Context) {},
	)

	handler(c)
	assert.Equal(t, `{"message":"Denied"}`, blw.MockBody.String())
	assert.Equal(t, http.StatusForbidden, blw.MockStatusCode)
}

func TestList(t *testing.T) {
	mockResponse := `{"meta":{"count":1},"results":[{"age":20,"name":"test"}]}`
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	blw := &MockGinBodyResponseWriter{MockBody: bytes.NewBufferString(""), ResponseWriter: c.Writer}
	c.Writer = blw

	basePath := "/objects"

	objectManager := &testObjectManager{}
	objectManager.Database = append(
		objectManager.Database,
		testObject{Pk: 1, Name: "test", Age: 20},
	)
	viewSet := NewViewSet[testObject, testObjectRequest](
		basePath, "/:pk", nil, nil, objectManager, nil, nil, nil, nil,
	)

	List(DEFAULT_LIST_ACTION, viewSet, c)

	assert.Equal(t, mockResponse, blw.MockBody.String())
	assert.Equal(t, http.StatusOK, blw.MockStatusCode)
}

func TestListWithGetObjectsError(t *testing.T) {
	mockResponse := `{"message":"Get objects error"}`
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	blw := &MockGinBodyResponseWriter{MockBody: bytes.NewBufferString(""), ResponseWriter: c.Writer}
	c.Writer = blw

	basePath := "/objects"

	objectManager := &testObjectManager{RaiseError: true}
	objectManager.Database = append(
		objectManager.Database,
		testObject{Pk: 1, Name: "test", Age: 20},
	)
	viewSet := NewViewSet[testObject, testObjectRequest](
		basePath, "/:pk", nil, nil, objectManager, nil, nil, nil, nil,
	)

	List(DEFAULT_LIST_ACTION, viewSet, c)

	assert.Equal(t, mockResponse, blw.MockBody.String())
	assert.Equal(t, http.StatusInternalServerError, blw.MockStatusCode)
}

func TestListWithSerializeError(t *testing.T) {
	mockResponse := `{"message":"Many serialize error"}`
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	blw := &MockGinBodyResponseWriter{MockBody: bytes.NewBufferString(""), ResponseWriter: c.Writer}
	c.Writer = blw

	basePath := "/objects"
	serializer := &MockSerializerAlwaysError[testObject]{}

	objectManager := &testObjectManager{}
	objectManager.Database = append(
		objectManager.Database,
		testObject{Pk: 1, Name: "test", Age: 20},
	)
	viewSet := NewViewSet[testObject, testObjectRequest](
		basePath, "/:pk", nil, nil, objectManager, nil, nil, serializer, nil,
	)

	List(DEFAULT_LIST_ACTION, viewSet, c)

	assert.Equal(t, mockResponse, blw.MockBody.String())
	assert.Equal(t, http.StatusInternalServerError, blw.MockStatusCode)
}

func TestRetrieve(t *testing.T) {
	mockResponse := `{"age":20,"name":"test"}`
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	blw := &MockGinBodyResponseWriter{MockBody: bytes.NewBufferString(""), ResponseWriter: c.Writer}
	c.Writer = blw
	c.Params = []gin.Param{
		{
			Key:   "pk",
			Value: "1",
		},
	}

	basePath := "/objects"

	objectManager := &testObjectManager{}
	objectManager.Database = append(
		objectManager.Database,
		testObject{Pk: 1, Name: "test", Age: 20},
	)
	viewSet := NewViewSet[testObject, testObjectRequest](
		basePath, "/:pk", nil, nil, objectManager, nil, nil, nil, nil,
	)

	Retrieve(DEFAULT_RETRIEVE_ACTION, viewSet, c)

	assert.Equal(t, mockResponse, blw.MockBody.String())
	assert.Equal(t, http.StatusOK, blw.MockStatusCode)
}

func TestRetrieveWithGetObjectError(t *testing.T) {
	mockResponse := `{"message":"Object not found"}`
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	blw := &MockGinBodyResponseWriter{MockBody: bytes.NewBufferString(""), ResponseWriter: c.Writer}
	c.Writer = blw
	c.Params = []gin.Param{
		{
			Key:   "pk",
			Value: "2",
		},
	}

	basePath := "/objects"

	objectManager := &testObjectManager{}
	objectManager.Database = append(
		objectManager.Database,
		testObject{Pk: 1, Name: "test", Age: 20},
	)
	viewSet := NewViewSet[testObject, testObjectRequest](
		basePath, "/:pk", nil, nil, objectManager, nil, nil, nil, nil,
	)

	Retrieve(DEFAULT_RETRIEVE_ACTION, viewSet, c)

	assert.Equal(t, mockResponse, blw.MockBody.String())
	assert.Equal(t, http.StatusNotFound, blw.MockStatusCode)
}

func TestRetrieveWithSerializeError(t *testing.T) {
	mockResponse := `{"message":"Serialize error"}`
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	blw := &MockGinBodyResponseWriter{MockBody: bytes.NewBufferString(""), ResponseWriter: c.Writer}
	c.Writer = blw
	c.Params = []gin.Param{
		{
			Key:   "pk",
			Value: "1",
		},
	}

	basePath := "/objects"
	serializer := &MockSerializerAlwaysError[testObject]{}

	objectManager := &testObjectManager{}
	objectManager.Database = append(
		objectManager.Database,
		testObject{Pk: 1, Name: "test", Age: 20},
	)
	viewSet := NewViewSet[testObject, testObjectRequest](
		basePath, "/:pk", nil, nil, objectManager, nil, nil, serializer, nil,
	)

	Retrieve(DEFAULT_RETRIEVE_ACTION, viewSet, c)

	assert.Equal(t, mockResponse, blw.MockBody.String())
	assert.Equal(t, http.StatusInternalServerError, blw.MockStatusCode)
}

func TestCreate(t *testing.T) {
	mockResponse := `{"age":21,"name":"test 2"}`
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	blw := &MockGinBodyResponseWriter{MockBody: bytes.NewBufferString(""), ResponseWriter: c.Writer}
	c.Writer = blw
	c.Request = &http.Request{
		Header: make(http.Header),
	}

	MockJsonPost(c, map[string]any{"name": "test 2", "age": 21})

	basePath := "/objects"

	objectManager := &testObjectManager{}
	objectManager.Database = append(
		objectManager.Database,
		testObject{Pk: 1, Name: "test", Age: 20},
	)
	viewSet := NewViewSet[testObject, testObjectRequest](
		basePath, "/:pk", nil, nil, objectManager, nil, nil, nil, nil,
	)

	Create(DEFAULT_CREATE_ACTION, viewSet, c)

	assert.Equal(t, mockResponse, blw.MockBody.String())
	assert.Equal(t, http.StatusCreated, blw.MockStatusCode)
}

func TestCreateWithValidateError(t *testing.T) {
	mockResponse := `{"message":"Key: 'testObjectRequest.Age' Error:Field validation for 'Age' failed on the 'required' tag"}`
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	blw := &MockGinBodyResponseWriter{MockBody: bytes.NewBufferString(""), ResponseWriter: c.Writer}
	c.Writer = blw
	c.Request = &http.Request{
		Header: make(http.Header),
	}

	MockJsonPost(c, map[string]any{"name": "test 2"})

	basePath := "/objects"

	objectManager := &testObjectManager{}
	objectManager.Database = append(
		objectManager.Database,
		testObject{Pk: 1, Name: "test", Age: 20},
	)
	viewSet := NewViewSet[testObject, testObjectRequest](
		basePath, "/:pk", nil, nil, objectManager, nil, nil, nil, nil,
	)

	Create(DEFAULT_CREATE_ACTION, viewSet, c)

	assert.Equal(t, mockResponse, blw.MockBody.String())
	assert.Equal(t, http.StatusBadRequest, blw.MockStatusCode)
}

func TestCreateWithSaveError(t *testing.T) {
	mockResponse := `{"message":"Saving error"}`
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	blw := &MockGinBodyResponseWriter{MockBody: bytes.NewBufferString(""), ResponseWriter: c.Writer}
	c.Writer = blw
	c.Request = &http.Request{
		Header: make(http.Header),
	}

	MockJsonPost(c, map[string]any{"name": "test 2", "age": 21})

	basePath := "/objects"

	objectManager := &testObjectManager{RaiseError: true}
	objectManager.Database = append(
		objectManager.Database,
		testObject{Pk: 1, Name: "test", Age: 20},
	)
	viewSet := NewViewSet[testObject, testObjectRequest](
		basePath, "/:pk", nil, nil, objectManager, nil, nil, nil, nil,
	)

	Create(DEFAULT_CREATE_ACTION, viewSet, c)

	assert.Equal(t, mockResponse, blw.MockBody.String())
	assert.Equal(t, http.StatusBadRequest, blw.MockStatusCode)
}

func TestCreateWithSerializeError(t *testing.T) {
	mockResponse := `{"message":"Serialize error"}`
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	blw := &MockGinBodyResponseWriter{MockBody: bytes.NewBufferString(""), ResponseWriter: c.Writer}
	c.Writer = blw
	c.Request = &http.Request{
		Header: make(http.Header),
	}

	MockJsonPost(c, map[string]any{"name": "test 2", "age": 21})

	basePath := "/objects"
	serializer := &MockSerializerAlwaysError[testObject]{}

	objectManager := &testObjectManager{}
	objectManager.Database = append(
		objectManager.Database,
		testObject{Pk: 1, Name: "test", Age: 20},
	)
	viewSet := NewViewSet[testObject, testObjectRequest](
		basePath, "/:pk", nil, nil, objectManager, nil, nil, serializer, nil,
	)

	Create(DEFAULT_CREATE_ACTION, viewSet, c)

	assert.Equal(t, mockResponse, blw.MockBody.String())
	assert.Equal(t, http.StatusInternalServerError, blw.MockStatusCode)
}

func TestUpdate(t *testing.T) {
	mockResponse := `{"age":21,"name":"test 2"}`
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	blw := &MockGinBodyResponseWriter{MockBody: bytes.NewBufferString(""), ResponseWriter: c.Writer}
	c.Writer = blw
	c.Request = &http.Request{
		Header: make(http.Header),
	}
	c.Params = []gin.Param{
		{
			Key:   "pk",
			Value: "1",
		},
	}

	MockJsonPut(c, map[string]any{"name": "test 2", "age": 21})

	basePath := "/objects"

	objectManager := &testObjectManager{}
	objectManager.Database = append(
		objectManager.Database,
		testObject{Pk: 1, Name: "test", Age: 20},
	)
	viewSet := NewViewSet[testObject, testObjectRequest](
		basePath, "/:pk", nil, nil, objectManager, nil, nil, nil, nil,
	)

	Update(DEFAULT_UPDATE_ACTION, viewSet, c)

	assert.Equal(t, mockResponse, blw.MockBody.String())
	assert.Equal(t, http.StatusOK, blw.MockStatusCode)
}

func TestUpdateWithGetObjectError(t *testing.T) {
	mockResponse := `{"message":"Object not found"}`
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	blw := &MockGinBodyResponseWriter{MockBody: bytes.NewBufferString(""), ResponseWriter: c.Writer}
	c.Writer = blw
	c.Request = &http.Request{
		Header: make(http.Header),
	}
	c.Params = []gin.Param{
		{
			Key:   "pk",
			Value: "2",
		},
	}

	MockJsonPut(c, map[string]any{"name": "test 2", "age": 21})

	basePath := "/objects"

	objectManager := &testObjectManager{}
	objectManager.Database = append(
		objectManager.Database,
		testObject{Pk: 1, Name: "test", Age: 20},
	)
	viewSet := NewViewSet[testObject, testObjectRequest](
		basePath, "/:pk", nil, nil, objectManager, nil, nil, nil, nil,
	)

	Update(DEFAULT_UPDATE_ACTION, viewSet, c)

	assert.Equal(t, mockResponse, blw.MockBody.String())
	assert.Equal(t, http.StatusNotFound, blw.MockStatusCode)
}

func TestUpdateWithValidateError(t *testing.T) {
	mockResponse := `{"message":"Key: 'testObjectRequest.Age' Error:Field validation for 'Age' failed on the 'required' tag"}`
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	blw := &MockGinBodyResponseWriter{MockBody: bytes.NewBufferString(""), ResponseWriter: c.Writer}
	c.Writer = blw
	c.Request = &http.Request{
		Header: make(http.Header),
	}
	c.Params = []gin.Param{
		{
			Key:   "pk",
			Value: "1",
		},
	}

	MockJsonPut(c, map[string]any{"name": "test 2"})

	basePath := "/objects"

	objectManager := &testObjectManager{}
	objectManager.Database = append(
		objectManager.Database,
		testObject{Pk: 1, Name: "test", Age: 20},
	)
	viewSet := NewViewSet[testObject, testObjectRequest](
		basePath, "/:pk", nil, nil, objectManager, nil, nil, nil, nil,
	)

	Update(DEFAULT_UPDATE_ACTION, viewSet, c)

	assert.Equal(t, mockResponse, blw.MockBody.String())
	assert.Equal(t, http.StatusBadRequest, blw.MockStatusCode)
}

func TestUpdateWithSaveError(t *testing.T) {
	mockResponse := `{"message":"Saving error"}`
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	blw := &MockGinBodyResponseWriter{MockBody: bytes.NewBufferString(""), ResponseWriter: c.Writer}
	c.Writer = blw
	c.Request = &http.Request{
		Header: make(http.Header),
	}
	c.Params = []gin.Param{
		{
			Key:   "pk",
			Value: "1",
		},
	}

	MockJsonPut(c, map[string]any{"name": "test 2", "age": 21})

	basePath := "/objects"

	objectManager := &testObjectManager{RaiseError: true}
	objectManager.Database = append(
		objectManager.Database,
		testObject{Pk: 1, Name: "test", Age: 20},
	)
	viewSet := NewViewSet[testObject, testObjectRequest](
		basePath, "/:pk", nil, nil, objectManager, nil, nil, nil, nil,
	)

	Update(DEFAULT_UPDATE_ACTION, viewSet, c)

	assert.Equal(t, mockResponse, blw.MockBody.String())
	assert.Equal(t, http.StatusBadRequest, blw.MockStatusCode)
}

func TestUpdateWithSerializeError(t *testing.T) {
	mockResponse := `{"message":"Serialize error"}`
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	blw := &MockGinBodyResponseWriter{MockBody: bytes.NewBufferString(""), ResponseWriter: c.Writer}
	c.Writer = blw
	c.Request = &http.Request{
		Header: make(http.Header),
	}
	c.Params = []gin.Param{
		{
			Key:   "pk",
			Value: "1",
		},
	}

	MockJsonPut(c, map[string]any{"name": "test 2", "age": 21})

	basePath := "/objects"
	serializer := &MockSerializerAlwaysError[testObject]{}

	objectManager := &testObjectManager{}
	objectManager.Database = append(
		objectManager.Database,
		testObject{Pk: 1, Name: "test", Age: 20},
	)
	viewSet := NewViewSet[testObject, testObjectRequest](
		basePath, "/:pk", nil, nil, objectManager, nil, nil, serializer, nil,
	)

	Update(DEFAULT_UPDATE_ACTION, viewSet, c)

	assert.Equal(t, mockResponse, blw.MockBody.String())
	assert.Equal(t, http.StatusInternalServerError, blw.MockStatusCode)
}

func TestDelete(t *testing.T) {
	mockResponse := ``
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	blw := &MockGinBodyResponseWriter{MockBody: bytes.NewBufferString(""), ResponseWriter: c.Writer}
	c.Writer = blw
	c.Request = &http.Request{
		Header: make(http.Header),
	}
	c.Params = []gin.Param{
		{
			Key:   "pk",
			Value: "1",
		},
	}
	c.Request.Method = "DELETE"

	basePath := "/objects"

	objectManager := &testObjectManager{}
	objectManager.Database = append(
		objectManager.Database,
		testObject{Pk: 1, Name: "test", Age: 20},
	)
	viewSet := NewViewSet[testObject, testObjectRequest](
		basePath, "/:pk", nil, nil, objectManager, nil, nil, nil, nil,
	)

	Delete(DEFAULT_DELETE_ACTION, viewSet, c)

	assert.Equal(t, mockResponse, blw.MockBody.String())
	assert.Equal(t, http.StatusNoContent, blw.MockStatusCode)
}

func TestDeleteWithGetObjectError(t *testing.T) {
	mockResponse := `{"message":"Object not found"}`
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	blw := &MockGinBodyResponseWriter{MockBody: bytes.NewBufferString(""), ResponseWriter: c.Writer}
	c.Writer = blw
	c.Request = &http.Request{
		Header: make(http.Header),
	}
	c.Params = []gin.Param{
		{
			Key:   "pk",
			Value: "2",
		},
	}
	c.Request.Method = "DELETE"

	basePath := "/objects"

	objectManager := &testObjectManager{}
	objectManager.Database = append(
		objectManager.Database,
		testObject{Pk: 1, Name: "test", Age: 20},
	)
	viewSet := NewViewSet[testObject, testObjectRequest](
		basePath, "/:pk", nil, nil, objectManager, nil, nil, nil, nil,
	)

	Delete(DEFAULT_DELETE_ACTION, viewSet, c)

	assert.Equal(t, mockResponse, blw.MockBody.String())
	assert.Equal(t, http.StatusNotFound, blw.MockStatusCode)
}

func TestDeleteWithDeleteError(t *testing.T) {
	mockResponse := `{"message":"Deleting error"}`
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	blw := &MockGinBodyResponseWriter{MockBody: bytes.NewBufferString(""), ResponseWriter: c.Writer}
	c.Writer = blw
	c.Request = &http.Request{
		Header: make(http.Header),
	}
	c.Params = []gin.Param{
		{
			Key:   "pk",
			Value: "1",
		},
	}
	c.Request.Method = "DELETE"

	basePath := "/objects"

	objectManager := &testObjectManager{RaiseError: true}
	objectManager.Database = append(
		objectManager.Database,
		testObject{Pk: 1, Name: "test", Age: 20},
	)
	viewSet := NewViewSet[testObject, testObjectRequest](
		basePath, "/:pk", nil, nil, objectManager, nil, nil, nil, nil,
	)

	Delete(DEFAULT_DELETE_ACTION, viewSet, c)

	assert.Equal(t, mockResponse, blw.MockBody.String())
	assert.Equal(t, http.StatusBadRequest, blw.MockStatusCode)
}
