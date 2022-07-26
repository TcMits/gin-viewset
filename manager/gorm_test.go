package manager

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"testing"

	"gorm.io/driver/postgres"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type person struct {
	ID   uint   `gorm:"primarykey" mapstructure:"-"`
	Name string `mapstructure:"name"`
}

type personRequest struct {
	Name string `mapstructure:"name"`
}

func (person) TableName() string {
	return "person"
}

type personURI struct {
	ID uint `uri:"pk" binding:"required" mapstructure:"id"`
}

type dbSuite struct {
	suite.Suite
	DB   *gorm.DB
	mock sqlmock.Sqlmock
}

func (s *dbSuite) SetupTest() {
	var (
		db  *sql.DB
		err error
	)

	db, s.mock, err = sqlmock.New()
	if err != nil {
		panic(err) // Error here
	}

	s.DB, err = gorm.Open(postgres.New(postgres.Config{
		Conn:                 db,
		PreferSimpleProtocol: true,
	}), &gorm.Config{
		SkipDefaultTransaction: true,
	})

	if err != nil {
		panic(err) // Error here
	}
}

func (s *dbSuite) TearDownTest() {
	db, err := s.DB.DB()
	if err != nil {
		return
	}
	defer db.Close()
}

func (s *dbSuite) TestNewGormManager() {
	gormManager := NewGormManager[person, personRequest, personURI](
		s.DB, nil, nil, nil, nil, "db", func(c *gin.Context) func(*gorm.DB) *gorm.DB {
			return func(d *gorm.DB) *gorm.DB { return d }
		},
	)

	assert.Equal(s.T(), s.DB, gormManager.db)
	assert.Equal(s.T(), "db", gormManager.ginContextKey)
	assert.Equal(s.T(), 1, len(gormManager.scopeGenerators))
	assert.NotZero(s.T(), gormManager.paginateFunc)
	assert.NotZero(s.T(), gormManager.performCreateFunc)
	assert.NotZero(s.T(), gormManager.performUpdateFunc)
	assert.NotZero(s.T(), gormManager.performDeleteFunc)
}

func (s *dbSuite) TestGormManagerNewDBWithContext() {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	gormManager := NewGormManager[person, personRequest, personURI](
		s.DB, nil, nil, nil, nil, "db",
	)

	newDB := gormManager.newDBWithContext(c)

	dbInContext, ok := c.Get("db")
	assert.Equal(s.T(), true, ok)
	assert.Equal(s.T(), newDB, dbInContext.(*gorm.DB))
}

func (s *dbSuite) TestGormManagerGetDBWithContext() {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	gormManager := NewGormManager[person, personRequest, personURI](
		s.DB, nil, nil, nil, nil, "db",
	)

	newDB := gormManager.GetDBWithContext(c)

	dbInContext, ok := c.Get("db")
	assert.Equal(s.T(), true, ok)
	assert.Equal(s.T(), newDB, dbInContext.(*gorm.DB))
}

func (s *dbSuite) TestGormManagerGetDBWithContextExistDB() {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	gormManager := NewGormManager[person, personRequest, personURI](
		s.DB, nil, nil, nil, nil, "db",
	)

	newDB1 := gormManager.newDBWithContext(c)
	newDB2 := gormManager.GetDBWithContext(c)

	dbInContext, ok := c.Get("db")
	assert.Equal(s.T(), true, ok)
	assert.Equal(s.T(), newDB1, dbInContext.(*gorm.DB))
	assert.Equal(s.T(), newDB2, dbInContext.(*gorm.DB))
}

func (s *dbSuite) TestGormManagerGetDBWithContextExistNil() {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	gormManager := NewGormManager[person, personRequest, personURI](
		s.DB, nil, nil, nil, nil, "db",
	)

	c.Set("db", nil)
	newDB := gormManager.GetDBWithContext(c)

	dbInContext, ok := c.Get("db")
	assert.Equal(s.T(), true, ok)
	assert.Equal(s.T(), newDB, dbInContext.(*gorm.DB))
}

func (s *dbSuite) TestGormManagerGetQuerySet() {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	callBack := ""

	gormManager := NewGormManager[person, personRequest, personURI](
		s.DB, nil, nil, nil, nil, "db", func(c *gin.Context) func(*gorm.DB) *gorm.DB {
			callBack = "test"
			return func(d *gorm.DB) *gorm.DB {
				return d
			}
		},
	)

	gormManager.GetQuerySet(c)
	assert.Equal(s.T(), "test", callBack)
}

func (s *dbSuite) TestDefaultPaginateFunc() {
	mockURL, _ := url.Parse("https://example.com/?limit=1&offset=0&with_count=true")
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = &http.Request{
		Header: make(http.Header),
		URL:    mockURL,
	}
	person1 := person{1, "phuc"}
	person2 := person{2, "huy"}

	gormManager := NewGormManager[person, personRequest, personURI](
		s.DB.Model(&person{}), nil, nil, nil, nil, "db",
	)

	entities := make([]*person, 0, 20)
	paginatedMeta := map[string]any{}

	s.mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT count(*) FROM "person"`),
	).WillReturnRows(
		sqlmock.NewRows([]string{"count(1)"}).AddRow(2),
	)
	s.mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT * FROM "person" LIMIT 2`),
	).WillReturnRows(
		sqlmock.NewRows([]string{"id", "name"}).AddRow(
			person1.ID, person1.Name,
		).AddRow(
			person2.ID, person2.Name,
		),
	)

	err := gormManager.GetObjects(&entities, &paginatedMeta, c)

	assert.NoError(s.T(), s.mock.ExpectationsWereMet())
	assert.Equal(s.T(), nil, err)
	assert.Equal(s.T(), 1, len(entities))
	assert.Equal(s.T(), "phuc", entities[0].Name)
	assert.Equal(s.T(), uint(1), entities[0].ID)
	assert.Equal(s.T(), "https://example.com/?limit=1&offset=1&with_count=true", paginatedMeta["next"])
	assert.Equal(s.T(), nil, paginatedMeta["previous"])
	assert.Equal(s.T(), int64(2), paginatedMeta["count"])
}

func (s *dbSuite) TestDefaultPaginateFuncWithBindError() {
	mockURL, _ := url.Parse("https://example.com/?limit=1&offset=-1&with_count=true")
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = &http.Request{
		Header: make(http.Header),
		URL:    mockURL,
	}

	gormManager := NewGormManager[person, personRequest, personURI](
		s.DB.Model(&person{}), nil, nil, nil, nil, "db",
	)

	entities := make([]*person, 0, 20)
	paginatedMeta := map[string]any{}

	err := gormManager.GetObjects(&entities, &paginatedMeta, c)

	assert.NotEqual(s.T(), nil, err)

}

func (s *dbSuite) TestDefaultPaginateFuncWithPrevious() {
	mockURL, _ := url.Parse("https://example.com/?limit=1&offset=1")
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = &http.Request{
		Header: make(http.Header),
		URL:    mockURL,
	}
	personIns := person{2, "huy"}

	gormManager := NewGormManager[person, personRequest, personURI](
		s.DB.Model(&person{}), nil, nil, nil, nil, "db",
	)

	entities := make([]*person, 0, 20)
	paginatedMeta := map[string]any{}

	s.mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT * FROM "person" LIMIT 2 OFFSET 1`),
	).WillReturnRows(
		sqlmock.NewRows([]string{"id", "name"}).AddRow(
			personIns.ID, personIns.Name,
		),
	)

	err := gormManager.GetObjects(&entities, &paginatedMeta, c)

	assert.NoError(s.T(), s.mock.ExpectationsWereMet())
	assert.Equal(s.T(), nil, err)
	assert.Equal(s.T(), 1, len(entities))
	assert.Equal(s.T(), "huy", entities[0].Name)
	assert.Equal(s.T(), uint(2), entities[0].ID)
	assert.Equal(s.T(), nil, paginatedMeta["next"])
	assert.Equal(s.T(), "https://example.com/?limit=1&offset=0", paginatedMeta["previous"])
}

func (s *dbSuite) TestDefaultPaginateFuncWithLargeLimit() {
	mockURL, _ := url.Parse("https://example.com/?limit=10&offset=1")
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = &http.Request{
		Header: make(http.Header),
		URL:    mockURL,
	}
	personIns := person{2, "huy"}

	gormManager := NewGormManager[person, personRequest, personURI](
		s.DB.Model(&person{}), nil, nil, nil, nil, "db",
	)

	entities := make([]*person, 0, 20)
	paginatedMeta := map[string]any{}

	s.mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT * FROM "person" LIMIT 11 OFFSET 1`),
	).WillReturnRows(
		sqlmock.NewRows([]string{"id", "name"}).AddRow(
			personIns.ID, personIns.Name,
		),
	)

	err := gormManager.GetObjects(&entities, &paginatedMeta, c)

	assert.NoError(s.T(), s.mock.ExpectationsWereMet())
	assert.Equal(s.T(), nil, err)
	assert.Equal(s.T(), 1, len(entities))
	assert.Equal(s.T(), "huy", entities[0].Name)
	assert.Equal(s.T(), uint(2), entities[0].ID)
	assert.Equal(s.T(), nil, paginatedMeta["next"])
	assert.Equal(s.T(), "https://example.com/?limit=10&offset=0", paginatedMeta["previous"])
}

func (s *dbSuite) TestDefaultPaginateFuncWithoutLimit() {
	mockURL, _ := url.Parse("https://example.com/")
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = &http.Request{
		Header: make(http.Header),
		URL:    mockURL,
	}
	personIns := person{2, "huy"}

	gormManager := NewGormManager[person, personRequest, personURI](
		s.DB.Model(&person{}), nil, nil, nil, nil, "db",
	)

	entities := make([]*person, 0, 20)
	paginatedMeta := map[string]any{}

	s.mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT * FROM "person" LIMIT 21`),
	).WillReturnRows(
		sqlmock.NewRows([]string{"id", "name"}).AddRow(
			personIns.ID, personIns.Name,
		),
	)

	err := gormManager.GetObjects(&entities, &paginatedMeta, c)

	assert.NoError(s.T(), s.mock.ExpectationsWereMet())
	assert.Equal(s.T(), nil, err)
	assert.Equal(s.T(), 1, len(entities))
	assert.Equal(s.T(), "huy", entities[0].Name)
	assert.Equal(s.T(), uint(2), entities[0].ID)
	assert.Equal(s.T(), nil, paginatedMeta["next"])
	assert.Equal(s.T(), nil, paginatedMeta["previous"])
}

func (s *dbSuite) TestGormManagerGetObject() {
	mockURL, _ := url.Parse("https://example.com/")
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = &http.Request{
		Header: make(http.Header),
		URL:    mockURL,
	}
	c.Params = []gin.Param{
		{
			Key:   "pk",
			Value: "1",
		},
	}
	personIns := person{1, "phuc"}

	gormManager := NewGormManager[person, personRequest, personURI](
		s.DB.Model(&person{}), nil, nil, nil, nil, "db",
	)

	s.mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT * FROM "person" WHERE "person"."id" = $1 ORDER BY "person"."id" LIMIT 1`),
	).WithArgs(1).WillReturnRows(
		sqlmock.NewRows([]string{"id", "name"}).AddRow(
			personIns.ID, personIns.Name,
		),
	)

	entity := new(person)

	err := gormManager.GetObject(&entity, c)

	assert.NoError(s.T(), s.mock.ExpectationsWereMet())
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), personIns.Name, entity.Name)
	assert.Equal(s.T(), personIns.ID, entity.ID)
}

func (s *dbSuite) TestGormManagerGetObjectWithBindingError() {
	mockURL, _ := url.Parse("https://example.com/")
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = &http.Request{
		Header: make(http.Header),
		URL:    mockURL,
	}
	c.Params = []gin.Param{
		{
			Key:   "id",
			Value: "1",
		},
	}

	gormManager := NewGormManager[person, personRequest, personURI](
		s.DB.Model(&person{}), nil, nil, nil, nil, "db",
	)

	entity := new(person)

	err := gormManager.GetObject(&entity, c)

	assert.Error(s.T(), err)
}

func (s *dbSuite) TestGormManagerGetObjectWithFirstError() {
	mockURL, _ := url.Parse("https://example.com/")
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = &http.Request{
		Header: make(http.Header),
		URL:    mockURL,
	}
	c.Params = []gin.Param{
		{
			Key:   "pk",
			Value: "2",
		},
	}

	gormManager := NewGormManager[person, personRequest, personURI](
		s.DB.Model(&person{}), nil, nil, nil, nil, "db",
	)

	s.mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT * FROM "person" WHERE "person"."id" = $1 ORDER BY "person"."id" LIMIT 1`),
	).WithArgs(2).WillReturnRows(
		sqlmock.NewRows([]string{"id", "name"}),
	)

	entity := new(person)

	err := gormManager.GetObject(&entity, c)

	assert.NoError(s.T(), s.mock.ExpectationsWereMet())
	assert.Error(s.T(), err)
}

func (s *dbSuite) TestGormManagerDefaultCreateFunc() {
	mockURL, _ := url.Parse("https://example.com/")
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = &http.Request{
		Header: make(http.Header),
		URL:    mockURL,
	}

	gormManager := NewGormManager[person, personRequest, personURI](
		s.DB.Model(&person{}), nil, nil, nil, nil, "db",
	)

	s.mock.ExpectQuery(regexp.QuoteMeta(
		`INSERT INTO "person" ("name") VALUES ($1) RETURNING "id"`),
	).WithArgs("phuc").WillReturnRows(
		sqlmock.NewRows([]string{"id"}).AddRow(1),
	)

	var entity *person
	validatedData := personRequest{Name: "phuc"}

	err := gormManager.Save(&entity, &validatedData, c)

	assert.NoError(s.T(), s.mock.ExpectationsWereMet())
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "phuc", entity.Name)
	assert.Equal(s.T(), uint(1), entity.ID)
}

func (s *dbSuite) TestGormManagerDefaultUpdateFunc() {
	mockURL, _ := url.Parse("https://example.com/")
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = &http.Request{
		Header: make(http.Header),
		URL:    mockURL,
	}

	gormManager := NewGormManager[person, personRequest, personURI](
		s.DB.Model(&person{}), nil, nil, nil, nil, "db",
	)

	s.mock.ExpectExec(
		`UPDATE "person"`,
	).WithArgs(
		"phuc 2", 1,
	).WillReturnResult(
		sqlmock.NewResult(1, 1),
	)

	entity := &person{ID: 1, Name: "phuc"}
	validatedData := personRequest{Name: "phuc 2"}

	err := gormManager.Save(&entity, &validatedData, c)

	assert.NoError(s.T(), s.mock.ExpectationsWereMet())
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "phuc 2", entity.Name)
	assert.Equal(s.T(), uint(1), entity.ID)
}

func (s *dbSuite) TestGormManagerDefaultDeleteFunc() {
	mockURL, _ := url.Parse("https://example.com/")
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = &http.Request{
		Header: make(http.Header),
		URL:    mockURL,
	}

	gormManager := NewGormManager[person, personRequest, personURI](
		s.DB.Model(&person{}), nil, nil, nil, nil, "db",
	)

	s.mock.ExpectExec(
		`DELETE FROM "person"`,
	).WithArgs(1).WillReturnResult(
		sqlmock.NewResult(1, 1),
	)

	entity := &person{ID: 1, Name: "phuc"}

	err := gormManager.Delete(&entity, c)
	assert.NoError(s.T(), s.mock.ExpectationsWereMet())
	assert.NoError(s.T(), err)
}

func TestGorm(t *testing.T) {
	suite.Run(t, &dbSuite{})
}
