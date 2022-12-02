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

    const image_bucket = new google.storageBucket.StorageBucket(this, 'image-bucket', {
      location: region,
      name: `image-bucket-${project}`,
    });

    new google.storageBucket.StorageBucket(this, 'result-bucket', {
      location: region,
      name: `result-bucket-${project}`,
    });

    const translate_topic = new google.pubsubTopic.PubsubTopic(this, 'translate-topic', {
      name: 'translate',
    });

    const result_topic = new google.pubsubTopic.PubsubTopic(this, 'result-topic', {
      name: 'result',
    });

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
      eventTrigger: {
        eventType: 'google.cloud.storage.object.v1.finalized',
        eventFilters: [{
          attribute: 'bucket',
          value: image_bucket.name,
        }],
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
      eventTrigger: {
        pubsubTopic: result_topic.name,
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
      eventTrigger: {
        pubsubTopic: translate_topic.name,
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
