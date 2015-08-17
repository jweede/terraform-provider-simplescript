# terraform-provider-simplescript

Terraform plugin for providing values via script.

Here's an example terraform:

    resource "simplescript_run" "test" {
        command = "echo '{"test": "123"}'"
    }
    
    output "cmd-output" {
        value = "${simplescript_run.test.text_output}"
    }
    
    output "cmd-json" {
        value = "${simplescript_run.test.json_output.test}"
    }
