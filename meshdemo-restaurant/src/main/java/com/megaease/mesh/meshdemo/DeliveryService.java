package com.megaease.mesh.meshdemo;

import org.springframework.cloud.openfeign.FeignClient;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestMethod;

@FeignClient("${easemesh.demo.remote-service:delivery-mesh}")
interface DeliveryService {
    @RequestMapping(name = "/", method = RequestMethod.POST, produces = "application/json", consumes = "application/json")
    DeliveryResponse getResponse(@RequestBody DeliveryRequest request);
}
