package component

import (
	"fmt"
	"testing"

	"github.com/base-org/pessimism/internal/core"
	"github.com/stretchr/testify/assert"
)

func Test_Add_Remove_Ingress(t *testing.T) {
	var tests = []struct {
		name        string
		description string

		constructionLogic func() *ingressHandler
		testLogic         func(*testing.T, *ingressHandler)
	}{
		{
			name:        "Successful Add Test",
			description: "When a register type is passed to AddIngress function, it should be successfully added to handler ingress mapping",

			constructionLogic: func() *ingressHandler {
				handler := newIngressHandler()
				return handler
			},

			testLogic: func(t *testing.T, ih *ingressHandler) {

				err := ih.createIngress(core.GethBlock)
				assert.NoError(t, err, "geth.block register should added as an egress")

			},
		},
		{
			name:        "Failed Add Test",
			description: "When the same register type is added twice to AddIngress function, the second add should fail with key collisions",

			constructionLogic: func() *ingressHandler {
				handler := newIngressHandler()
				if err := handler.createIngress(core.GethBlock); err != nil {
					panic(err)
				}

				return handler
			},

			testLogic: func(t *testing.T, ih *ingressHandler) {
				err := ih.createIngress(core.GethBlock)

				assert.Error(t, err, "geth.block register should fail to be added")
				assert.Equal(t, err.Error(), fmt.Sprintf(ingressAlreadyExistsErr, core.GethBlock.String()))

			},
		},
	}

	for i, tc := range tests {
		t.Run(fmt.Sprintf("%d-%s", i, tc.name), func(t *testing.T) {
			testIngress := tc.constructionLogic()
			tc.testLogic(t, testIngress)
		})

	}
}
