package google

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"google.golang.org/api/spanner/v1"
)

type SpannerInstanceOperationWaiter struct {
	Service *spanner.Service
	Op      *spanner.Operation
}

func (w *SpannerInstanceOperationWaiter) Conf(timeoutMinutes int) *resource.StateChangeConf {
	return &resource.StateChangeConf{
		Pending:    []string{"false"},
		Target:     []string{"true"},
		Refresh:    w.RefreshFunc(),
		Timeout:    time.Duration(timeoutMinutes) * time.Minute,
		MinTimeout: 2 * time.Second,
	}
}

func (w *SpannerInstanceOperationWaiter) RefreshFunc() resource.StateRefreshFunc {
	return func() (interface{}, string, error) {

		op, err := w.Service.Projects.Instances.Operations.Get(w.Op.Name).Do()

		if err != nil {
			return nil, "", err
		}

		log.Printf("[DEBUG] Got %v while polling for operation %s's 'done' status", op.Done, w.Op.Name)

		return op, fmt.Sprint(op.Done), nil
	}
}

func spannerInstanceOperationWait(config *Config, op *spanner.Operation, activity string, timeoutMin int) error {
	if op.Done {
		if op.Error != nil {
			return fmt.Errorf("Error code %v, message: %s", op.Error.Code, op.Error.Message)
		}
		return nil
	}

	w := &SpannerInstanceOperationWaiter{
		Service: config.clientSpanner,
		Op:      op,
	}

	opRaw, err := w.Conf(timeoutMin).WaitForState()
	if err != nil {
		return fmt.Errorf("Error waiting for %s: %s", activity, err)
	}

	op = opRaw.(*spanner.Operation)
	if op.Error != nil {
		return fmt.Errorf("Error code %v, message: %s", op.Error.Code, op.Error.Message)
	}

	return nil

}
