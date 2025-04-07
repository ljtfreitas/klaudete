import kcl_plugin.inventory

_environments = inventory.list_resources("environment")

environments = [{id: e.id, nurn: e.metadata.nurn, alias: e.metadata.alias} for e in _environments]
