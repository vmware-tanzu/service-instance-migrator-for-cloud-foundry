/*
 *  Copyright 2022 VMware, Inc.
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *  http://www.apache.org/licenses/LICENSE-2.0
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */

package om

import (
	"context"
	"fmt"
	"strings"

	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/exec"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/log"
)

func Run(e exec.Executor, ctx context.Context, data interface{}, args ...string) (exec.Result, error) {
	foundation, ok := data.(config.OpsManager)
	if !ok {
		log.Fatal("failed to convert type to Foundation")
	}
	var lines []string

	if len(args) > 0 {
		lines = []string{
			fmt.Sprintf(`OM_CLIENT_ID='%s' OM_CLIENT_SECRET='%s' OM_USERNAME='%s' OM_PASSWORD='%s' om -t '%s' -k %s`,
				foundation.ClientID,
				foundation.ClientSecret,
				foundation.Username,
				foundation.Password,
				foundation.URL,
				joinArgsBySpace(args)),
		}
	} else {
		lines = []string{
			fmt.Sprintf(`echo "export OM_TARGET='%s'"`, foundation.URL),
			fmt.Sprintf(`echo "export OM_CLIENT_ID='%s'"`, foundation.ClientID),
			fmt.Sprintf(`echo "export OM_CLIENT_SECRET='%s'"`, foundation.ClientSecret),
			fmt.Sprintf(`echo "export OM_USERNAME='%s'"`, foundation.Username),
			fmt.Sprintf(`echo "export OM_PASSWORD='%s'"`, foundation.Password),
		}
	}

	return e.Execute(ctx, strings.NewReader(strings.Join(lines, "\n")))
}

func joinArgsBySpace(args []string) string {
	arg := make([]string, 0, len(args))
	arg = append(arg, args...)
	return strings.Join(arg, " ")
}
