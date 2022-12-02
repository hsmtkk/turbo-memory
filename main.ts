// Copyright (c) HashiCorp, Inc
// SPDX-License-Identifier: MPL-2.0
import { Construct } from "constructs";
import { App, TerraformStack, CloudBackend, NamedCloudWorkspace } from "cdktf";
import * as google from '@cdktf/provider-google';
//import * as path from 'path';

const project = 'turbo-memory';
const region = 'asia-northeast1';
//const repository = 'turbo-memory';

class MyStack extends TerraformStack {
  constructor(scope: Construct, id: string) {
    super(scope, id);

    new google.provider.GoogleProvider(this, 'google', {
      project,
      region,
    });

    const buckets = ['image', 'result']
    for(const bucket of buckets){
      new google.storageBucket.StorageBucket(this, `${bucket}-bucket`, {
        location: region,
        name: `${bucket}-${project}`,
      });
    }

  }
}

const app = new App();
const stack = new MyStack(app, "turbo-memory");
new CloudBackend(stack, {
  hostname: "app.terraform.io",
  organization: "hsmtkkdefault",
  workspaces: new NamedCloudWorkspace("turbo-memory")
});
app.synth();
