package viewset

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
    LIST_ACTION = iota
    RETRIEVE_ACTION
    CREATE_ACTION
    UPDATE_ACTION
    DESTROY_ACTION
)



type ObjectManager interface {
    GetObjects(*gin.Context) (interface{}, interface{})
    GetObject(string, *gin.Context) (interface{}, error)
    Save(interface{}, *gin.Context) error
    Remove(interface{}, *gin.Context) error
}


type Serializer interface {
    Serialize(interface{}, *gin.Context) (interface{})
    ManySerialize(interface{}, *gin.Context) (interface{})
    GetValidatedData(interface{}, *gin.Context) (interface{}, error)
}


type viewSet struct {
    baseURL string
    detailURLLookup string
    objectManager ObjectManager
    serializer Serializer
    excludeAction []int
}


func Register(
    handler *gin.RouterGroup,
    baseURL string, // "/users"
    detailURLLookup string,  // "pk"
    objectManager ObjectManager,
    serializer Serializer,
    excludeAction []int,
) {
	r := &viewSet{
        objectManager: objectManager,
        serializer: serializer,
        baseURL: baseURL,
        detailURLLookup: detailURLLookup,
        excludeAction: excludeAction,
    }
    r.asView(handler)
}

func (vs *viewSet) asView(handler *gin.RouterGroup) {
    h := handler.Group(vs.baseURL)
	{
        if vs.shouldRegister(LIST_ACTION) {
            h.GET("/", vs.list)
        }
        if vs.shouldRegister(CREATE_ACTION) {
            h.POST("/", vs.create)
        }
        if vs.shouldRegister(RETRIEVE_ACTION) {
            h.GET("/:" + vs.detailURLLookup + "/", vs.retrive)
        }
        if vs.shouldRegister(UPDATE_ACTION) {
            h.PUT("/:" + vs.detailURLLookup + "/", vs.update)
            h.PATCH("/:" + vs.detailURLLookup + "/", vs.partialUpdate)
        }
        if vs.shouldRegister(DESTROY_ACTION) {
            h.DELETE("/:" + vs.detailURLLookup + "/", vs.destroy)
        }
	}
}


func (vs *viewSet) shouldRegister(action int) bool {
    for _, a := range vs.excludeAction {
        if a == action {
            return false
        }
    }
    return true
}


func (vs *viewSet) list(c *gin.Context) {
    objects, meta := vs.objectManager.GetObjects(c)
    results := vs.serializer.ManySerialize(objects, c)
	c.JSON(http.StatusOK, gin.H{
        "meta": meta,
        "results": results,
    })
}


func (vs *viewSet) retrive(c *gin.Context) {
    object, err := vs.objectManager.GetObject(vs.detailURLLookup, c)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
        return
    }
    result := vs.serializer.Serialize(object, c)
	c.JSON(http.StatusOK, result)
}


func (vs *viewSet) create(c *gin.Context) {
    object, err := vs.serializer.GetValidatedData(nil, c)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
        return
    }

    if err := vs.objectManager.Save(object, c); err != nil {
        c.JSON(http.StatusNotAcceptable, gin.H{"message": err.Error()})
        return
    }
    result := vs.serializer.Serialize(object, c)
	c.JSON(http.StatusCreated, result)
}


func (vs *viewSet) update(c *gin.Context) {
    object, err := vs.objectManager.GetObject(vs.detailURLLookup, c)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
        return
    }

    validatedObject, err := vs.serializer.GetValidatedData(object, c)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
        return
    }

    if err := vs.objectManager.Save(validatedObject, c); err != nil {
        c.JSON(http.StatusNotAcceptable, gin.H{"message": err.Error()})
        return
    }
    result := vs.serializer.Serialize(validatedObject, c)
	c.JSON(http.StatusOK, result)
}

func (vs *viewSet) partialUpdate(c *gin.Context) {
    vs.update(c)
}


func (vs *viewSet) destroy(c *gin.Context) {
    object, err := vs.objectManager.GetObject(vs.detailURLLookup, c)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
        return
    }
    if err := vs.objectManager.Remove(object, c); err != nil {
        c.JSON(http.StatusNotAcceptable, gin.H{"message": err.Error()})
        return
    }
    c.Status(http.StatusNoContent)
}

