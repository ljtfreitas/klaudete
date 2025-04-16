import kcl_plugin.inventory

_accounts = ["nurn:nu:infra:aws:aws-account/nu-isc-ch-prod", \
                    "nurn:nu:infra:aws:aws-account/nu-isc-br-prod"]

_query = "in(operates-on) from(aws-iam-platform)"

accounts = [inventory.get_resource(a, _query) for a in _accounts]

# environments = inventory.resources_by_type(account, "aws-iam-platform")

# nurn = "nurn:nu:infra:aws-security-group/sg-demo-orchestrator-cap2"

# query = "out(is-deployed-on) to(aws-account) also out(is-deployed-on) to(environment)"

# sg = inventory.get_resource(nurn, query)

# _environments = inventory.list_resources("environment")

# _environments = [{id: e.id, nurn: e.metadata.nurn, alias: e.metadata.alias} for e in inventory.list_resources("environment")]

# _aws_accounts = [{account:  {id: a.id, nurn: a.metadata.nurn, alias: a.metadata.alias}, environment: e} for e in _environments for a in inventory.list_resources("aws-account")]

# output = _aws_accounts