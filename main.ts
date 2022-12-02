// Copyright (c) HashiCorp, Inc
// SPDX-License-Identifier: MPL-2.0
import { Construct } from "constructs";
import { App, TerraformStack, CloudBackend, NamedCloudWorkspace, TerraformAsset, AssetType } from "cdktf";
import * as google from '@cdktf/provider-google';
import * as path from 'path';

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

    const services = ['cloudfunctions'];
    for(const service of services){
      new google.projectService.ProjectService(this, `enable-${service}`, {
        service: `${service}.googleapis.com`,
      });
    }

    const buckets = ['image', 'result']
    for(const bucket of buckets){
      new google.storageBucket.StorageBucket(this, `${bucket}-bucket`, {
        location: region,
        name: `${bucket}-${project}`,
      });
    }

    const topics = ['translate', 'result']
    for(const topic of topics){
      new google.pubsubTopic.PubsubTopic(this, `${topic}-topic`, {
        name: topic,
      });
    }

    const source_bucket = new google.storageBucket.StorageBucket(this, 'source-bucket', {
      location: region,
      name: `source-${project}`,      
    });

    // extract

    const extract_asset = new TerraformAsset(this, 'extract-asset', {
      path: path.resolve('extract'),
      type: AssetType.ARCHIVE,
    });

    const extract_object = new google.storageBucketObject.StorageBucketObject(this, 'extract-object', {
      bucket: source_bucket.name,
      name: `${extract_asset.assetHash}.zip`,
      source: extract_asset.path,
    });

    new google.cloudfunctions2Function.Cloudfunctions2Function(this, 'extract-function', {
      buildConfig: {
        entryPoint: 'extract',
        runtime: 'go119',
        source: {
          storageSource: {
            bucket: source_bucket.name,
            object: extract_object.name,
          },          
        },
      },
      location: region,
      name: 'ocr-extract',
      serviceConfig: {
        environmentVariables: {
          'GCP_PROJECT': project,
          'TRANSLATE_TOPIC': 'translate',
          'RESULT_TOPIC': 'result',
          'TO_LANG': 'es,en,fr,ja',
        },
        minInstanceCount: 0,
        maxInstanceCount: 1,
      },
    });

    // save

    const save_asset = new TerraformAsset(this, 'save-asset', {
      path: path.resolve('save'),
      type: AssetType.ARCHIVE,
    });

    const save_object = new google.storageBucketObject.StorageBucketObject(this, 'save-object', {
      bucket: source_bucket.name,
      name: `${save_asset.assetHash}.zip`,
      source: save_asset.path,
    });

    new google.cloudfunctions2Function.Cloudfunctions2Function(this, 'save-function', {
      buildConfig: {
        entryPoint: 'save',
        runtime: 'go119',
        source: {
          storageSource: {
            bucket: source_bucket.name,
            object: save_object.name,
          },          
        },
      },
      location: region,
      name: 'ocr-save',
      serviceConfig: {
        environmentVariables: {
          'GCP_PROJECT': project,
          'RESULT_BUCKET': 'result',
        },
        minInstanceCount: 0,
        maxInstanceCount: 1,
      },
    });
    
    // translate

    const translate_asset = new TerraformAsset(this, 'translate-asset', {
      path: path.resolve('translate'),
      type: AssetType.ARCHIVE,
    });

    const translate_object = new google.storageBucketObject.StorageBucketObject(this, 'translate-object', {
      bucket: source_bucket.name,
      name: `${translate_asset.assetHash}.zip`,
      source: translate_asset.path,
    });

    new google.cloudfunctions2Function.Cloudfunctions2Function(this, 'translate-function', {
      buildConfig: {
        entryPoint: 'translate',
        runtime: 'go119',
        source: {
          storageSource: {
            bucket: source_bucket.name,
            object: translate_object.name,
          },          
        },
      },
      location: region,
      name: 'ocr-translate',
      serviceConfig: {
        environmentVariables: {
          'GCP_PROJECT': project,
          'RESULT_TOPIC': 'result',
        },
        minInstanceCount: 0,
        maxInstanceCount: 1,
      },
    });

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
