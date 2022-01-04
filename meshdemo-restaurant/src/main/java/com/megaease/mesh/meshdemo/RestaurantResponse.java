package com.megaease.mesh.meshdemo;

import com.fasterxml.jackson.databind.PropertyNamingStrategy;
import com.fasterxml.jackson.databind.annotation.JsonNaming;
import com.fasterxml.jackson.annotation.JsonInclude;
import lombok.Builder;
import lombok.Data;

@Data
@Builder
@JsonNaming(PropertyNamingStrategy.SnakeCaseStrategy.class)
public class RestaurantResponse {
    private String orderId;
    private String food;
    private String deliveryTime;

    @JsonInclude(JsonInclude.Include.NON_EMPTY)
    private String coupon;
}
