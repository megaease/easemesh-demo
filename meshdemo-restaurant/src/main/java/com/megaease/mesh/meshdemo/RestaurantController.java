package com.megaease.mesh.meshdemo;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RestController;

import javax.annotation.Resource;

@RestController
public class RestaurantController {
    private final static Logger LOGGER = LoggerFactory.getLogger(RestaurantController.class);
    @Resource
    DeliveryService deliveryService;

    @PostMapping("/")
    public RestaurantResponse order(@RequestBody OrderRequest order) {
        DeliveryRequest req = new DeliveryRequest(order.getOrderID(), order.getFood());
        LOGGER.warn("send request {} to delivery service add consume and produce and requestbody sr8 ribbon ", req);
        DeliveryResponse resp = deliveryService.getResponse(req);
        LOGGER.warn("received response {} from delivery service", resp);

        RestaurantResponse restaurantResp = RestaurantResponse.builder().orderId(order.getOrderID())
                .food(resp.getItem()).deliveryTime(resp.getDeliveryTime()).build();

        // For canary Beijing.
        if ("restaurant-mesh-beijing".equals(System.getenv("SERVICE_NAME"))) {
            String deliveryTime = restaurantResp.getDeliveryTime() + " (cook duration: 5m)";
            restaurantResp.setDeliveryTime(deliveryTime);
        }

        // For canary Android.
        if (resp.getLate() != null && resp.getLate().booleanValue() == true) {
            restaurantResp.setCoupon("$5");
        }

        return restaurantResp;
    }

    @PostMapping("/mock")
    public RestaurantResponse mock(@RequestBody OrderRequest order) {
        return RestaurantResponse.builder().orderId(order.getOrderID()).food(order.getFood())
                .deliveryTime("2021-10-02T16:00:00").build();
    }
}
