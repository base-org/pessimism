package registry

import (
	"context"
	"log"
	"math/big"
	"time"

	"github.com/base-org/pessimism/internal/config"
	"github.com/base-org/pessimism/internal/etl/component"
	"github.com/base-org/pessimism/internal/models"
	"github.com/ethereum/go-ethereum/ethclient"
)

const (
	pollInterval = 200
)

// GethBlockODef ...GethBlock register oracle definition used to drive oracle component
type GethBlockODef struct {
	cfg *config.OracleConfig
	// TODO - Bind this to an interface and mock so that this logic can be tested
	client *ethclient.Client

	currHeight *big.Int
}

// NewGethBlockOracle ... Initializer
func NewGethBlockOracle(ctx context.Context,
	ot models.PipelineType, cfg *config.OracleConfig) (component.Component, error) {
	od := &GethBlockODef{cfg: cfg, currHeight: nil}

	return component.NewOracle(ctx, ot, od)
}

func (oracle *GethBlockODef) ConfigureRoutine() error {
	// TODO - Introduce starting block parameter
	log.Print("Setting up GETH Block client")
	client, err := ethclient.Dial(oracle.cfg.RPCEndpoint)
	if err != nil {
		return err
	}

	oracle.client = client
	return nil
}

// BackTestRoutine ...
func (oracle *GethBlockODef) BackTestRoutine(_ context.Context, _ chan models.TransitData) error {
	// TODO - implement

	return nil
}

// ReadRoutine ... Sequentially polls go-ethereum compatible execution
// client using monotonic block height variable for block metadata
// & writes block metadata to output listener components
func (oracle *GethBlockODef) ReadRoutine(ctx context.Context, componentChan chan models.TransitData) error {
	// NOTE - This poller logic is really bad and doesn't
	//        currently compensate for a lot of edge cases, some of the obvious being:
	// 1 - Client timeouts/failures; ie embed retry logic
	// 2 - Only reads most recent headers and doesn't take a starting block # - DONE
	// 3 - No optionality support for an ending block

	ticker := time.NewTicker(pollInterval * time.Millisecond)

	for {
		select {
		case <-ticker.C:
			height := oracle.currHeight
			header, err := oracle.client.HeaderByNumber(ctx, height)
			if err != nil {
				// log.Printf("Header fetching error: %s", err.Error())
				continue
			}

			block, err := oracle.client.BlockByNumber(ctx, header.Number)
			if err != nil {
				// log.Printf("Error fetching block @ height %d: %s", height, err)
				continue
			}

			log.Printf("%d", height)

			// TODO - Add support for database persistence

			log.Printf("Writing to component channel")
			componentChan <- models.TransitData{
				Timestamp: time.Now(),
				Type:      GethBlock,
				Value:     *block,
			}

			if height != nil {
				// height += 1
				height.Add(height, big.NewInt(1))
			} else {
				height = header.Number
			}

			oracle.currHeight = height

		case <-ctx.Done():
			return nil
		}
	}
}
