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

    const result_bucket = new google.storageBucket.StorageBucket(this, 'result-bucket', {
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

    const storage_service_account = new google.dataGoogleStorageProjectServiceAccount.DataGoogleStorageProjectServiceAccount(this, 'storage-service-account');

    new google.projectIamMember.ProjectIamMember(this, 'storage-account-pubsub', {
      member: `serviceAccount:${storage_service_account.emailAddress}`,
      project,
      role: 'roles/pubsub.publisher',
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
        triggerRegion: region,
      },
      location: region,
      name: 'ocr-extract',
      serviceConfig: {
        environmentVariables: {
          'GCP_PROJECT': project,
          'TRANSLATE_TOPIC': translate_topic.name,
          'RESULT_TOPIC': result_topic.name,
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
        eventType: 'google.cloud.pubsub.topic.v1.messagePublished',
        pubsubTopic: result_topic.id,
      },
      location: region,
      name: 'ocr-save',
      serviceConfig: {
        environmentVariables: {
          'GCP_PROJECT': project,
          'RESULT_BUCKET': result_bucket.name,
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
        eventType: 'google.cloud.pubsub.topic.v1.messagePublished',
        pubsubTopic: translate_topic.id,
      },
      location: region,
      name: 'ocr-translate',
      serviceConfig: {
        environmentVariables: {
          'GCP_PROJECT': project,
          'RESULT_TOPIC': result_topic.name,
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
