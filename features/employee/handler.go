package employee

import (
	"care-cordination/lib/middleware"
	"care-cordination/lib/resp"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type EmployeeHandler struct {
	employeeService EmployeeService
	mdw             *middleware.Middleware
}

func NewEmployeeHandler(
	employeeService EmployeeService,
	mdw *middleware.Middleware,
) *EmployeeHandler {
	return &EmployeeHandler{
		employeeService: employeeService,
		mdw:             mdw,
	}
}

func (h *EmployeeHandler) SetupEmployeeRoutes(router *gin.Engine) {
	employee := router.Group("/employees")
	employee.Use(h.mdw.AuthMdw())

	employee.GET("/me", h.GetMyProfile)
	employee.GET("", h.mdw.PaginationMdw(), h.ListEmployees)
	employee.GET("/:id", h.GetEmployeeByID)
	employee.POST("", h.mdw.RequirePermission("employee", "write"), h.CreateEmployee)
	employee.PUT("/:id", h.mdw.RequirePermission("employee", "write"), h.UpdateEmployee)
	employee.DELETE("/:id", h.mdw.RequirePermission("employee", "delete"), h.DeleteEmployee)
}

// @Summary Create an employee
// @Description Create a new employee
// @Tags Employee
// @Accept json
// @Produce json
// @Param employee body CreateEmployeeRequest true "Employee"
// @Success 200 {object} resp.SuccessResponse[CreateEmployeeResponse]
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
	ctx.JSON(http.StatusOK, resp.Success(result, "Employee created successfully"))
}

// @Summary List employees
// @Description List all employees with pagination and search
// @Tags Employee
// @Accept json
// @Produce json
// @Param page query int false "Page number (default: 1)"
// @Param page_size query int false "Page size (default: 10, max: 100)"
// @Param search query string false "Search by first name, last name, or full name"
// @Success 200 {object} resp.SuccessResponse[resp.PaginationResponse[[]ListEmployeesResponse]]
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
	ctx.JSON(http.StatusOK, resp.Success(result, "Employees listed successfully"))
}

// @Summary Get employee by ID
// @Description Get a single employee by their ID with all details
// @Tags Employee
// @Produce json
// @Param id path string true "Employee ID"
// @Success 200 {object} resp.SuccessResponse[GetEmployeeByIDResponse]
// @Failure 400 {object} resp.ErrorResponse
// @Failure 401 {object} resp.ErrorResponse
// @Failure 404 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /employees/{id} [get]
func (h *EmployeeHandler) GetEmployeeByID(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, resp.Error(ErrInvalidRequest))
		return
	}

	result, err := h.employeeService.GetEmployeeByID(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, ErrInternal):
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		default:
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		}
		return
	}
	ctx.JSON(http.StatusOK, resp.Success(result, "Employee retrieved successfully"))
}

// @Summary Get logged-in user's employee profile
// @Description Get the employee profile of the currently authenticated user
// @Tags Employee
// @Produce json
// @Success 200 {object} resp.SuccessResponse[GetMyProfileResponse]
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /employees/me [get]
func (h *EmployeeHandler) GetMyProfile(ctx *gin.Context) {
	result, err := h.employeeService.GetMyProfile(ctx)
	if err != nil {
		switch {
		case errors.Is(err, ErrInternal):
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		default:
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		}
		return
	}
	ctx.JSON(http.StatusOK, resp.Success(result, "Employee profile retrieved successfully"))
}

// @Summary Update an employee
// @Description Update an existing employee's details including password
// @Tags Employee
// @Accept json
// @Produce json
// @Param id path string true "Employee ID"
// @Param employee body UpdateEmployeeRequest true "Employee update data"
// @Success 200 {object} resp.SuccessResponse[UpdateEmployeeResponse]
// @Failure 400 {object} resp.ErrorResponse
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /employees/{id} [put]
func (h *EmployeeHandler) UpdateEmployee(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, resp.Error(ErrInvalidRequest))
		return
	}

	var req UpdateEmployeeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, resp.Error(ErrInvalidRequest))
		return
	}

	result, err := h.employeeService.UpdateEmployee(ctx, id, &req)
	if err != nil {
		switch {
		case errors.Is(err, ErrInternal):
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		default:
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		}
		return
	}
	ctx.JSON(http.StatusOK, resp.Success(result, "Employee updated successfully"))
}

// @Summary Delete an employee
// @Description Soft delete an employee by ID
// @Tags Employee
// @Produce json
// @Param id path string true "Employee ID"
// @Success 200 {object} resp.SuccessResponse[any]
// @Failure 400 {object} resp.ErrorResponse
// @Failure 401 {object} resp.ErrorResponse
// @Failure 500 {object} resp.ErrorResponse
// @Router /employees/{id} [delete]
func (h *EmployeeHandler) DeleteEmployee(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, resp.Error(ErrInvalidRequest))
		return
	}

	err := h.employeeService.DeleteEmployee(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, ErrInternal):
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		default:
			ctx.JSON(http.StatusInternalServerError, resp.Error(err))
		}
		return
	}
	ctx.JSON(http.StatusOK, resp.Success(struct{}{}, "Employee deleted successfully"))
}
