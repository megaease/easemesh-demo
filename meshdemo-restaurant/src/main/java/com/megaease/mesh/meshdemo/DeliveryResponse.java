package com.megaease.mesh.meshdemo;

import com.fasterxml.jackson.databind.PropertyNamingStrategy;
import com.fasterxml.jackson.databind.annotation.JsonNaming;
import lombok.Data;

@Data
@JsonNaming(PropertyNamingStrategy.SnakeCaseStrategy.class)
public class DeliveryResponse {
    private String orderId;
    private String item;
    private String deliveryTime;

    private Boolean late;

    private String[] serviceTracings;
}
