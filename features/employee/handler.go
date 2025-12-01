package employee

import (
	"care-cordination/features/middleware"
	"care-cordination/lib/resp"
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

	employee.POST("", h.mdw.AuthMiddleware(), h.CreateEmployee)
	employee.GET("", h.mdw.AuthMiddleware(), h.ListEmployees)
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
// @Description List all employees
// @Tags Employee
// @Accept json
// @Produce json
// @Success 200 {object} []ListEmployeesResponse
// @Failure 400 {object} resp.ErrorResponse
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /employees [get]
func (h *EmployeeHandler) ListEmployees(ctx *gin.Context) {
	result, err := h.employeeService.ListEmployees(ctx)
	if err != nil {
		switch err {
		case ErrInternal:
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		default:
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		}
		return
	}
	ctx.JSON(http.StatusOK, result)
}
