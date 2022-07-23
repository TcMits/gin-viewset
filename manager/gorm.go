package manager

import (
	"strconv"

	"github.com/TcMits/viewset/pkg/urlclone"
	"github.com/gin-gonic/gin"
	"github.com/mitchellh/mapstructure"
	"gorm.io/gorm"
)

type ScopeGenerator func(c *gin.Context) func(*gorm.DB) *gorm.DB
type PaginateFunc[EntityType any] func(*[]*EntityType, *map[string]any, *gorm.DB, *gin.Context) error
type CreateFunc[EntityType any] func(**EntityType, *map[string]any, *gorm.DB, *gin.Context) error
type UpdateFunc[EntityType any] func(**EntityType, *map[string]any, *gorm.DB, *gin.Context) error
type DeleteFunc[EntityType any] func(**EntityType, *gorm.DB, *gin.Context) error

type LimitOffsetPaginator struct {
	Offset    uint `form:"offset"`
	Limit     uint `form:"limit"`
	WithCount bool `form:"withCount"`
}

type GormManager[EntityType any, URIType any] struct {
	db                *gorm.DB
	scopeGenerators   []ScopeGenerator
	paginateFunc      PaginateFunc[EntityType]
	performCreateFunc CreateFunc[EntityType]
	performUpdateFunc UpdateFunc[EntityType]
	performDeleteFunc DeleteFunc[EntityType]
	ginContextKey     string
}

func NewGormManager[EntityType any, URIType any](
	db *gorm.DB,
	paginateFunc PaginateFunc[EntityType],
	performCreateFunc CreateFunc[EntityType],
	performUpdateFunc UpdateFunc[EntityType],
	performDeleteFunc DeleteFunc[EntityType],
	ginContextKey string,
	scopeGenerators ...ScopeGenerator,
) *GormManager[EntityType, URIType] {
	if paginateFunc == nil {
		paginateFunc = DefaultPaginateFunc[EntityType]
	}
	if performCreateFunc == nil {
		performCreateFunc = DefaultCreateFunc[EntityType]
	}
	if performUpdateFunc == nil {
		performUpdateFunc = DefaultUpdateFunc[EntityType]
	}
	if performDeleteFunc == nil {
		performDeleteFunc = DefaultDeleteFunc[EntityType]
	}

	return &GormManager[EntityType, URIType]{
		db:                db,
		paginateFunc:      paginateFunc,
		scopeGenerators:   scopeGenerators,
		ginContextKey:     ginContextKey,
		performCreateFunc: performCreateFunc,
		performUpdateFunc: performUpdateFunc,
		performDeleteFunc: performDeleteFunc,
	}
}

func (manager *GormManager[_, _]) newDBWithContext(c *gin.Context) *gorm.DB {
	newDB := manager.db.WithContext(c)
	c.Set(manager.ginContextKey, newDB)
	return newDB
}

func (manager *GormManager[_, _]) GetDBWithContext(c *gin.Context) *gorm.DB {
	db, ok := c.Get(manager.ginContextKey)
	if !ok {
		return manager.newDBWithContext(c)
	}
	newDB, ok := db.(*gorm.DB)
	if !ok {
		return manager.newDBWithContext(c)
	}
	return newDB
}

func (manager *GormManager[_, _]) GetQuerySet(c *gin.Context) *gorm.DB {
	lenGen := len(manager.scopeGenerators)
	scopeFunctions := make([]func(*gorm.DB) *gorm.DB, 0, lenGen)
	for _, gen := range manager.scopeGenerators {
		scopeFunctions = append(scopeFunctions, gen(c))
	}
	return manager.GetDBWithContext(c).Scopes(scopeFunctions...)
}

func (manager *GormManager[EntityType, _]) GetObjects(
	dest *[]*EntityType, paginatedMeta *map[string]any, c *gin.Context) error {
	db := manager.GetQuerySet(c)
	return manager.paginateFunc(dest, paginatedMeta, db, c)
}

func (manager *GormManager[EntityType, URIType]) GetObject(
	dest **EntityType, c *gin.Context) error {
	*dest = new(EntityType)
	paramsValidator := new(URIType)
	filterKwargs := map[string]any{}
	if err := c.ShouldBindUri(paramsValidator); err != nil {
		return err
	}
	if err := mapstructure.Decode(paramsValidator, &filterKwargs); err != nil {
		return err
	}
	result := manager.GetQuerySet(c).First(*dest, filterKwargs)
	if err := result.Error; err != nil {
		return err
	}
	return nil
}

func (manager *GormManager[EntityType, _]) Save(
	dest **EntityType, validatedData *map[string]any, c *gin.Context) error {
	db := manager.GetDBWithContext(c)
	if *dest == nil {
		return manager.performCreateFunc(dest, validatedData, db, c)
	}
	return manager.performUpdateFunc(dest, validatedData, db, c)
}

func (manager *GormManager[EntityType, _]) Delete(
	dest **EntityType, c *gin.Context) error {
	db := manager.GetDBWithContext(c)
	return manager.performDeleteFunc(dest, db, c)
}

func DefaultPaginateFunc[EntityType any](
	dest *[]*EntityType, paginatedMeta *map[string]any, db *gorm.DB, c *gin.Context) error {
	// NOTE: using limit offset
	counter := 0
	paginator := LimitOffsetPaginator{Limit: 20}
	*paginatedMeta = map[string]any{}

	if err := c.ShouldBindQuery(&paginator); err != nil {
		return err
	}
	limit := int(paginator.Limit)
	offset := int(paginator.Offset)
	{
		if paginator.WithCount {
			countEntities := new(int64)
			if err := db.Count(countEntities).Error; err != nil {
				return err
			}
			(*paginatedMeta)["count"] = *countEntities
		}
		rows, err := db.Limit(limit + 1).Offset(offset).Rows()
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			counter += 1
			if counter > limit {
				break
			}
			entity := new(EntityType)
			db.ScanRows(rows, entity)
			*dest = append(*dest, entity)
		}
		if counter > limit {
			// have next
			nextURL := urlclone.CloneURL(c.Request.URL)
			values := nextURL.Query()
			values.Set("offset", strconv.Itoa(offset+limit))
			nextURL.RawQuery = values.Encode()
			(*paginatedMeta)["next"] = nextURL.String()
		} else {
			(*paginatedMeta)["next"] = nil
		}
		if offset > 0 {
			// have previous
			previousOffset := offset - limit
			if previousOffset < 0 {
				previousOffset = 0
			}
			previousURL := urlclone.CloneURL(c.Request.URL)
			values := previousURL.Query()
			values.Set("offset", strconv.Itoa(previousOffset))
			previousURL.RawQuery = values.Encode()
			(*paginatedMeta)["previous"] = previousURL.String()
		} else {
			(*paginatedMeta)["previous"] = nil
		}
	}
	return nil
}

func DefaultCreateFunc[EntityType any](dest **EntityType, validatedData *map[string]any, db *gorm.DB, _ *gin.Context) error {
	// NOTE: When creating from map, hooks won’t be invoked, associations won’t be saved and primary key values won’t be back filled
	*dest = new(EntityType)
	if err := mapstructure.Decode(validatedData, *dest); err != nil {
		return err
	}
	if err := db.Create(*dest).Error; err != nil {
		return err
	}
	return nil
}

func DefaultUpdateFunc[EntityType any](dest **EntityType, validatedData *map[string]any, db *gorm.DB, _ *gin.Context) error {
	if err := db.Model(*dest).Updates(validatedData).Error; err != nil {
		return err
	}
	return nil
}

func DefaultDeleteFunc[EntityType any](dest **EntityType, db *gorm.DB, _ *gin.Context) error {
	if err := db.Delete(*dest).Error; err != nil {
		return err
	}
	return nil
}
