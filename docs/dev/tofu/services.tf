# ---------------------------------------------------------------------------
# Demo services
#
# The UI shows a logo next to a service when its metadata carries an
# `external-source` value it knows (see
# ui/packages/sabo-ui/app/helpers/service/external-source.js). The services
# below stand in for external products that register themselves: OpenBao,
# OpenTofu and OpenWonton.
#
# They live on their own node (`demo-external`) on purpose. `consul_service`
# writes straight into the catalog, and a service parked on a node that runs an
# agent is removed again by the next anti-entropy sync, because that agent does
# not know about it. A node without an agent is also exactly the setup the
# provider's external-service flow expects.
# ---------------------------------------------------------------------------

resource "consul_node" "demo" {
  name    = "demo-external"
  address = "127.0.0.1"
}

resource "consul_service" "demo" {
  for_each = var.demo_services

  name    = "demo-${each.key}"
  node    = consul_node.demo.name
  address = "127.0.0.1"
  port    = each.value

  meta = {
    external-source = each.key
  }

  lifecycle {
    # The provider claims the `external-source` key for itself (it writes
    # "terraform" there and strips the key when reading a service back), so a
    # plain `meta` block reports a change on every plan that re-applying never
    # settles. The value is still written on create.
    ignore_changes = [meta]
  }
}
