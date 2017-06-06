// Copyright 2017 VMware, Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// This test suite tries to add, remove and replace vm to the _DEFAULT vmgroup
// Expected behavior is that add/rm/replace vm for _DEFAULT vmgroup should fail

package e2e

import (
	"log"
	"os"

	con "github.com/vmware/docker-volume-vsphere/tests/constants/admincli"
	"github.com/vmware/docker-volume-vsphere/tests/utils/admincli"
	"github.com/vmware/docker-volume-vsphere/tests/utils/dockercli"
	"github.com/vmware/docker-volume-vsphere/tests/utils/govc"
	"github.com/vmware/docker-volume-vsphere/tests/utils/inputparams"
	"github.com/vmware/docker-volume-vsphere/tests/utils/verification"
	. "gopkg.in/check.v1"
)

type VsanPolicyTestSuite struct {
	esxIP    string
	hostName string
}

func (s *VsanPolicyTestSuite) SetUpSuite(c *C) {
	log.Printf("VM1 IP: %s", os.Getenv("VM1"))
	s.hostName = govc.RetrieveVMNameFromIP(os.Getenv("VM1"))
	log.Printf("s.hostName=%s", s.hostName)
	s.esxIP = inputparams.GetEsxIP()
	out, err := admincli.ConfigInit(s.esxIP)
	c.Assert(err, IsNil, Commentf(out))
}

func (s *VsanPolicyTestSuite) TearDownSuite(c *C) {
	out, err := admincli.ConfigRemove(s.esxIP)
	c.Assert(err, IsNil, Commentf(out))
}

var _ = Suite(&VsanPolicyTestSuite{})

// Test step:
// 1. create a vsan policy
// 2. run "vmdkops_admin policy ls", check the "Active" colume of the output to make sure it
// is shown as "Unused"
// 3. create a volume on vsanDatastore with the vsan policy we created
// 4. run "docker volume inspect" on the volume to verify the output "vsan-policy-name" field
// 5. run "vmdkops_admin policy ls", check the "Active" colume of the output to make sure it
// is shown as "In use by 1 volumes"
// 6. run "vmdkops_admin policy rm" to remove the policy, which should fail since the volume still
// use the vsan policy
func (s *VsanPolicyTestSuite) TestDeleteVsanPolicyAlreadyInUse(c *C) {
	log.Printf("START: VsanPolicyTest.TestDeleteVsanPolicyAlreadyInUse")

	out, err := admincli.CreateVsanPolicy(s.esxIP, con.PolicyName, con.PolicyContent)
	c.Assert(err, IsNil, Commentf(out))

	/*	out, err = admincli.GetVsanPolicyListOutput(s.esxIP, policyName)
		log.Printf("VsanPolicyListOutput for policy %s is %s", policyName, out)
	*/

	res := admincli.VerifyActiveFromVsanPolicyListOutput(s.esxIP, con.PolicyName, "Unused")
	c.Assert(res, Equals, true, Commentf("vsanPolicy[%s] should be \"Unused\""))

	log.Printf("s.hostName=%s", s.hostName)
	out, err = dockercli.CreateVolumeWithVsanPolicy(s.hostName, con.VsanVol1, con.PolicyName)
	c.Assert(err, IsNil, Commentf(out))

	policyName := verification.GetVsanPolicyNameVolumeUsedDockerCli(con.VsanVol1, s.hostName)
	c.Assert(policyName, Equals, con.PolicyName, Commentf("The name of vsan policy used by volume "+con.VsanVol1+" returns incorrect value "+policyName))
	res = admincli.VerifyActiveFromVsanPolicyListOutput(s.esxIP, con.PolicyName, "In use by 1 volumes")
	c.Assert(res, Equals, true, Commentf("vsanPolicy[%s] should be \"In use by 1 volumes\""))

	out, err = admincli.RemoveVsanPolicy(s.esxIP, con.PolicyName)
	c.Assert(err, IsNil, Commentf(out))
	log.Printf("Remove vsanPolicy \"%s\" returns with %s", con.PolicyName, out)

	log.Printf("END: VsanPolicyTest.TestDeleteVsanPolicyAlreadyInUse")

}
