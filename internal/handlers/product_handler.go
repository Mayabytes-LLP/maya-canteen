package handlers

import (
	"maya-canteen/internal/database"
	"maya-canteen/internal/errors"
	"maya-canteen/internal/handlers/common"
	"maya-canteen/internal/models"
	"net/http"

	"github.com/gorilla/mux"
)

type ProductHandler struct {
	common.BaseHandler
}

func NewProductHandler(db database.Service) *ProductHandler {
	return &ProductHandler{
		BaseHandler: common.NewBaseHandler(db),
	}
}

func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var product models.Product
	if err := h.DecodeJSON(r, &product); err != nil {
		h.HandleError(w, err)
		return
	}

	if err := h.DB.CreateProduct(&product); err != nil {
		h.HandleError(w, errors.Internal(err))
		return
	}

	common.RespondWithSuccess(w, http.StatusCreated, product)
}

func (h *ProductHandler) GetAllProducts(w http.ResponseWriter, r *http.Request) {
	products, err := h.DB.GetAllProducts()
	if err != nil {
		h.HandleError(w, errors.Internal(err))
		return
	}

	common.RespondWithSuccess(w, http.StatusOK, products)
}

func (h *ProductHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := h.ParseID(vars, "id")
	if err != nil {
		h.HandleError(w, err)
		return
	}

	product, err := h.DB.GetProduct(id)
	if err != nil {
		h.HandleError(w, errors.Internal(err))
		return
	}

	if product == nil {
		h.HandleError(w, errors.NotFound("Product", id))
		return
	}

	common.RespondWithSuccess(w, http.StatusOK, product)
}

func (h *ProductHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := h.ParseID(vars, "id")
	if err != nil {
		h.HandleError(w, err)
		return
	}

	var product models.Product
	if err := h.DecodeJSON(r, &product); err != nil {
		h.HandleError(w, err)
		return
	}
	product.ID = id

	if err := h.DB.UpdateProduct(&product); err != nil {
		h.HandleError(w, errors.Internal(err))
		return
	}

	common.RespondWithSuccess(w, http.StatusOK, product)
}

func (h *ProductHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := h.ParseID(vars, "id")
	if err != nil {
		h.HandleError(w, err)
		return
	}

	if err := h.DB.DeleteProduct(id); err != nil {
		h.HandleError(w, errors.Internal(err))
		return
	}

	common.RespondWithSuccess(w, http.StatusNoContent, nil)
}
