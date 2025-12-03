package employee

import (
	"care-cordination/features/middleware"
	"care-cordination/lib/resp"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type EmployeeHandler struct {
	employeeService EmployeeService
	mdw             *middleware.Middleware
}

func NewEmployeeHandler(employeeService EmployeeService, mdw *middleware.Middleware) *EmployeeHandler {
	return &EmployeeHandler{
		employeeService: employeeService,
		mdw:             mdw,
	}
}

func (h *EmployeeHandler) SetupEmployeeRoutes(router *gin.Engine) {
	employee := router.Group("/employees")

	employee.POST("", h.mdw.AuthMdw(), h.CreateEmployee)
	employee.GET("", h.mdw.AuthMdw(), h.mdw.PaginationMdw(), h.ListEmployees)
}

// @Summary Create an employee
// @Description Create a new employee
// @Tags Employee
// @Accept json
// @Produce json
// @Param employee body CreateEmployeeRequest true "Employee"
// @Success 200 {object} CreateEmployeeResponse
// @Failure 400 {object} resp.ErrorResponse
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /employees [post]
func (h *EmployeeHandler) CreateEmployee(ctx *gin.Context) {
	var req CreateEmployeeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, resp.Error(ErrInvalidRequest))
		return
	}
	result, err := h.employeeService.CreateEmployee(ctx, &req)
	if err != nil {
		switch err {
		case ErrInvalidRequest:
			ctx.JSON(http.StatusBadRequest, resp.Error(err))
		case ErrInternal:
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		default:
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		}
		return
	}
	ctx.JSON(http.StatusOK, result)
}

// @Summary List employees
// @Description List all employees with pagination and search
// @Tags Employee
// @Accept json
// @Produce json
// @Param page query int false "Page number (default: 1)"
// @Param page_size query int false "Page size (default: 10, max: 100)"
// @Param search query string false "Search by first name, last name, or full name"
// @Success 200 {object} resp.PaginationResponse[[]ListEmployeesResponse]
// @Failure 400 {object} resp.ErrorResponse
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /employees [get]
func (h *EmployeeHandler) ListEmployees(ctx *gin.Context) {
	var req ListEmployeesRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, resp.Error(ErrInvalidRequest))
		return
	}
	result, err := h.employeeService.ListEmployees(ctx, &req)
	if err != nil {
		switch {
		case errors.Is(err, ErrInternal):
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		default:
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		}
		return
	}
	ctx.JSON(http.StatusOK, result)
}
