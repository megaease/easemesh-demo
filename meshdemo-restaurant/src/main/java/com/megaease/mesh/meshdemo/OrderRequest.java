package com.megaease.mesh.meshdemo;

import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.Data;

@Data
public class OrderRequest {
    @JsonProperty("order_id")
    private String orderID;
    private String food;
}
