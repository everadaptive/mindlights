import { EventBus } from "@vmw/transport/bus.api";
import { FabricApi } from "@vmw/transport/fabric.api";
import { Logger } from "@vmw/transport/log";
import { BusUtil } from "@vmw/transport/util/bus.util";
import React from "react";

export abstract class BaseComponent<A = any, S = any> extends React.Component<
  A,
  S
> {
  // get a reference to the bus, fabric and logger
  public bus: EventBus = BusUtil.getBusInstance();
  public fabric: FabricApi = this.bus.fabric;
  public log: Logger = this.bus.logger;
}
