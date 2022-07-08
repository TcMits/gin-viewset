package viewset

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)


type testObject struct {
    Pk string `json:"pk" form:"pk"`
    Name string `json:"name" form:"name"`
	Address string `json:"address" form:"address"`
}

type testObjectResponse struct {
    Pk string `json:"pk" form:"pk"`
    Name string `json:"name" form:"name"`
	Address string `json:"address" form:"address"`
    AdditionalData string `json:"additional_data"`
}

type testObjectRequest struct {
    Name string `json:"name" form:"name" binding:"required"`
	Address string `json:"address" form:"address" binding:"required"`
}

type testObjectManager struct {
    objects []testObject
    couter int
}

type testSerializer struct {}



func (om *testObjectManager) GetObjects(
    c *gin.Context,
) (interface{}, interface{}) {
    objects := &[]*testObject{}
    for i := range om.objects {
        *objects = append(*objects, &om.objects[i])
    }
    return objects, map[string]interface{}{
        "count": len(*objects),
    }
}


func (om *testObjectManager) GetObject(
    detailURLLookup string,
    c *gin.Context,
) (interface{}, error) {
    lookup := c.Param(detailURLLookup)
    for i, obj := range om.objects {
        if obj.Pk == lookup {
            return &om.objects[i], nil
        }
    }
    return nil, fmt.Errorf("Object not found")
}


func (om *testObjectManager) Save(object interface{}, c *gin.Context) error {
    o, ok := object.(*testObject)
    if !ok {
        return fmt.Errorf("Object can't update")
    }
    if o.Pk == "" {
        // create
        om.couter += 1
        o.Pk = strconv.Itoa(om.couter)
        om.objects = append(om.objects, *o)
    } else {
        for i, obj := range om.objects {
            if obj.Pk == o.Pk {
                om.objects[i].Name = o.Name
                om.objects[i].Address = o.Address
                break
            }
        }
    }
    return nil
}


func (om *testObjectManager) Remove(object interface{}, c *gin.Context) error {
    o, ok := object.(*testObject)
    if !ok {
        return fmt.Errorf("Object can't delete")
    }

    for i, obj := range om.objects {
        if o.Pk == obj.Pk {
            om.objects = append(om.objects[:i], om.objects[i + 1:]...)
            om.couter -= 1
            return nil
        }
    }

    return fmt.Errorf("Object not found")
}

func (ser *testSerializer) Serialize(
    object interface{},
    c *gin.Context,
) (interface{}) {
    d := &testObjectResponse{}
    o, okO := object.(*testObject)
    if !okO {
        return nil
    }
    d.Pk = o.Pk
    d.Name = o.Name
    d.Address = o.Address
    d.AdditionalData = "testing"
    return d
}


func (ser *testSerializer) ManySerialize(
    objects interface{},
    c *gin.Context,
) (interface{}) {
    objectsResponse := &[]interface{}{}
    objs, ok := objects.(*[]*testObject)
    if !ok {return &[]*testObjectResponse{}}
    for i := range *objs {
        *objectsResponse = append(
            *objectsResponse, 
            ser.Serialize((*objs)[i], c),
        )
    }
    return objectsResponse
}


func (ser *testSerializer) GetValidatedData(
    object interface{}, c *gin.Context,
) (interface{}, error) {
    d := &testObject{}
    if object != nil {
        // copy current opject to update
        o, okO := object.(*testObject)
        if !okO {
            return nil, fmt.Errorf("Invalid")
        }
        d.Pk = o.Pk
        d.Name = o.Name
        d.Address = o.Address
    }
    request := testObjectRequest{}
    if err := c.ShouldBindJSON(&request); err != nil {
        return nil, err
    }
    d.Name = request.Name
    d.Address = request.Address
    return d, nil
}


func SetUpRouter() *gin.Engine{
    gin.SetMode(gin.TestMode)
    router := gin.Default()
    return router
}

func TestRegisterListingAction(t *testing.T) {
    mockResponse := `{"meta":{"count":0},"results":[]}`
    var objectManager ObjectManager = &testObjectManager{}
    var serializer Serializer = &testSerializer{}
    r := SetUpRouter()
    rg := r.Group("/api")
    {
        Register(rg, "/test-objects", "pk", objectManager, serializer, []int{})
    }

    req, _ := http.NewRequest("GET", "/api/test-objects/", nil)
    w := httptest.NewRecorder()
    r.ServeHTTP(w, req)

    responseData, _ := ioutil.ReadAll(w.Body)
    assert.Equal(t, mockResponse, string(responseData))
    assert.Equal(t, http.StatusOK, w.Code)
}



func TestRegisterListingActionWithObjects(t *testing.T) {
    mockResponse := `{"meta":{"count":1},"results":[{"pk":"1","name":"test","address":"test","additional_data":"testing"}]}`
    objects := []testObject{}
    objects = append(objects, testObject{ Pk: "1", Name: "test", Address: "test", })
    var objectManager ObjectManager = &testObjectManager{
        objects: objects,
        couter: 1,
    }
    var serializer Serializer = &testSerializer{}
    r := SetUpRouter()
    rg := r.Group("/api")
    {
        Register(rg, "/test-objects", "pk", objectManager, serializer, []int{})
    }

    req, _ := http.NewRequest("GET", "/api/test-objects/", nil)
    w := httptest.NewRecorder()
    r.ServeHTTP(w, req)

    responseData, _ := ioutil.ReadAll(w.Body)
    assert.Equal(t, mockResponse, string(responseData))
    assert.Equal(t, http.StatusOK, w.Code)
}


func TestRegisterRetrieveAction(t *testing.T) {
    mockResponse := `{"message":"Object not found"}`
    var objectManager ObjectManager = &testObjectManager{}
    var serializer Serializer = &testSerializer{}
    r := SetUpRouter()
    rg := r.Group("/api")
    {
        Register(rg, "/test-objects", "pk", objectManager, serializer, []int{})
    }

    req, _ := http.NewRequest("GET", "/api/test-objects/1/", nil)
    w := httptest.NewRecorder()
    r.ServeHTTP(w, req)

    responseData, _ := ioutil.ReadAll(w.Body)
    assert.Equal(t, mockResponse, string(responseData))
    assert.Equal(t, http.StatusNotFound, w.Code)
}


func TestRegisterRetrieveActionWithObject(t *testing.T) {
    mockResponse := `{"pk":"1","name":"test","address":"test","additional_data":"testing"}`
    objects := []testObject{}
    objects = append(objects, testObject{ Pk: "1", Name: "test", Address: "test", })
    var objectManager ObjectManager = &testObjectManager{
        objects: objects,
        couter: 1,
    }
    var serializer Serializer = &testSerializer{}
    r := SetUpRouter()
    rg := r.Group("/api")
    {
        Register(rg, "/test-objects", "pk", objectManager, serializer, []int{})
    }

    req, _ := http.NewRequest("GET", "/api/test-objects/1/", nil)
    w := httptest.NewRecorder()
    r.ServeHTTP(w, req)

    responseData, _ := ioutil.ReadAll(w.Body)
    assert.Equal(t, mockResponse, string(responseData))
    assert.Equal(t, http.StatusOK, w.Code)
}



func TestRegisterCreateAction(t *testing.T) {
    mockResponse := `{"pk":"2","name":"test 2","address":"test 2","additional_data":"testing"}`
    objects := []testObject{}
    objects = append(objects, testObject{ Pk: "1", Name: "test", Address: "test", })
    var objectManager ObjectManager = &testObjectManager{
        objects: objects,
        couter: 1,
    }
    var serializer Serializer = &testSerializer{}
    r := SetUpRouter()
    rg := r.Group("/api")
    {
        Register(rg, "/test-objects", "pk", objectManager, serializer, []int{})
    }
    payload, _ := json.Marshal(map[string]interface{}{
        "name": "test 2",
        "address": "test 2",
    })

    req, _ := http.NewRequest("POST", "/api/test-objects/", bytes.NewBuffer(payload))
    w := httptest.NewRecorder()
    r.ServeHTTP(w, req)

    responseData, _ := ioutil.ReadAll(w.Body)
    assert.Equal(t, mockResponse, string(responseData))
    assert.Equal(t, http.StatusCreated, w.Code)
}


func TestRegisterUpdateActionWithObject(t *testing.T) {
    mockResponse := `{"pk":"1","name":"test testing","address":"test testing","additional_data":"testing"}`
    objects := []testObject{}
    objects = append(objects, testObject{ Pk: "1", Name: "test", Address: "test", })
    var objectManager ObjectManager = &testObjectManager{
        objects: objects,
        couter: 1,
    }
    var serializer Serializer = &testSerializer{}
    r := SetUpRouter()
    rg := r.Group("/api")
    {
        Register(rg, "/test-objects", "pk", objectManager, serializer, []int{})
    }
    payload, _ := json.Marshal(map[string]interface{}{
        "name": "test testing",
        "address": "test testing",
    })

    req, _ := http.NewRequest("PUT", "/api/test-objects/1/", bytes.NewBuffer(payload))
    w := httptest.NewRecorder()
    r.ServeHTTP(w, req)

    responseData, _ := ioutil.ReadAll(w.Body)
    assert.Equal(t, mockResponse, string(responseData))
    assert.Equal(t, http.StatusOK, w.Code)
}


func TestRegisterPartialUpdateAction(t *testing.T) {
    mockResponse := `{"pk":"1","name":"test testing","address":"test testing","additional_data":"testing"}`
    objects := []testObject{}
    objects = append(objects, testObject{ Pk: "1", Name: "test", Address: "test", })
    var objectManager ObjectManager = &testObjectManager{
        objects: objects,
        couter: 1,
    }
    var serializer Serializer = &testSerializer{}
    r := SetUpRouter()
    rg := r.Group("/api")
    {
        Register(rg, "/test-objects", "pk", objectManager, serializer, []int{})
    }
    payload, _ := json.Marshal(map[string]interface{}{
        "name": "test testing",
        "address": "test testing",
    })

    req, _ := http.NewRequest("PATCH", "/api/test-objects/1/", bytes.NewBuffer(payload))
    w := httptest.NewRecorder()
    r.ServeHTTP(w, req)

    responseData, _ := ioutil.ReadAll(w.Body)
    assert.Equal(t, mockResponse, string(responseData))
    assert.Equal(t, http.StatusOK, w.Code)
}


func TestRegisterPartialDestroyAction(t *testing.T) {
    objects := []testObject{}
    objects = append(objects, testObject{ Pk: "1", Name: "test", Address: "test", })
    var objectManager ObjectManager = &testObjectManager{
        objects: objects,
        couter: 1,
    }
    var serializer Serializer = &testSerializer{}
    r := SetUpRouter()
    rg := r.Group("/api")
    {
        Register(rg, "/test-objects", "pk", objectManager, serializer, []int{})
    }

    req, _ := http.NewRequest("DELETE", "/api/test-objects/1/", nil)
    w := httptest.NewRecorder()
    r.ServeHTTP(w, req)

    assert.Equal(t, http.StatusNoContent, w.Code)
}

